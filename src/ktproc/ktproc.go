package ktproc

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	s "strings"
	"time"
	"io/ioutil"
	"strconv"
	"context"

	config "mycs/src/kaoconfig"
	databasepool "mycs/src/kaodatabasepool"
)

var procCnt int


func KtProcess(user_id string, ctx context.Context, acc int) {
	config.Stdlog.Println(user_id, "Kt Process 시작 됨.")
	procCnt = 0
	for {
		if procCnt < 5 {
			select {
			case <-ctx.Done():
				config.Stdlog.Println(user_id, " - Ktxro process가 10초 후에 종료 됨.")
				time.Sleep(10 * time.Second)
				config.Stdlog.Println(user_id, " - Ktxro process 종료 완료")
				return
			default:

				var count sql.NullInt64
				tickSql := `
				select
					length(msgid) as cnt
				from
					DHN_RESULT dr
				where
					dr.result = 'P'
					and dr.send_group is null
					and ifnull(dr.reserve_dt, '00000000000000') <= date_format(now(), '%Y%m%d%H%i%S')
					and userid = ?
				limit 1
					`
				cnterr := databasepool.DB.QueryRow(tickSql, user_id).Scan(&count)

				if cnterr != nil && cnterr != sql.ErrNoRows {
					config.Stdlog.Println("DHN_RESULT Table - select 오류 : " + cnterr.Error())
					time.Sleep(10 * time.Second)
				} else {
					if count.Int64 > 0 {
						var startNow = time.Now()
						var group_no = fmt.Sprintf("%02d%02d%02d%02d%06d", startNow.Day(), startNow.Hour(), startNow.Minute(), startNow.Second(), (startNow.Nanosecond() / 1000))

						upError := updateReqeust(group_no, user_id)
						if upError != nil {
							config.Stdlog.Println(user_id, "Group No Update 오류", group_no)
						} else {
							go resProcess(group_no, user_id, acc)

						}
					}
				}
			}

		}
	}
}

func updateReqeust(group_no string, user_id string) error {

	tx, err := databasepool.DB.Begin()
	if err != nil {
		config.Stdlog.Println(err)
		return err
	}

	defer func() error {
		if err != nil {
			tx.Rollback()
			return err
		}
		err = tx.Commit()
		return err
	}()

	config.Stdlog.Println(user_id, "- KT 크로샷 Group No Update 시작", group_no)

	gudQuery := `
	update DHN_RESULT dr
	set	send_group = ?
	where result = 'P'
	  and send_group is null
	  and ifnull(reserve_dt, '00000000000000') <= date_format(now(), '%Y%m%d%H%i%S')
	  and userid = ?
	LIMIT 500
	`
	_, err = tx.Exec(gudQuery, group_no, user_id)

	if err != nil {
		config.Stdlog.Println(user_id, "- Group NO Update - Select error : ( group_no : "+group_no+" / user_id : "+user_id+" ) : "+err.Error())
		config.Stdlog.Println(gudQuery)
		return err
	}

	return nil
}

func resProcess(group_no string, user_id string, acc int) {
	defer func(){
		if r := recover(); r != nil {
			config.Stdlog.Println("KTX resProcess panic 발생 원인 : ", r)
			procCnt--
			if err, ok := r.(error); ok {
				if s.Contains(err.Error(), "connection refused") {
					for {
						config.Stdlog.Println("KTX resProcess send ping to DB")
						err := databasepool.DB.Ping()
						if err == nil {
							break
						}
						time.Sleep(10 * time.Second)
					}
				}
			}
		}
	}()
	procCnt++
	myacc := account[acc]
	client := NewMessage(myacc["apiKey"], myacc["apiPw"], myacc["userKey"], false, 3)
	
	var db = databasepool.DB
	var stdlog = config.Stdlog

	var msgid, code, message, message_type, msg_sms, phn, remark1, remark2, result, sms_lms_tit, sms_kind, sms_sender, res_dt, reserve_dt, mms_file1, mms_file2, mms_file3, userid, sms_len_check sql.NullString
	var msgLen sql.NullInt64
	var phnstr string

	var resquery = `
	SELECT
		msgid, 
		code, 
		message, 
		message_type, 
		(case when sms_kind = 'S' then 
			substr(convert(REMOVE_WS(msg_sms) using euckr),1,100)
		 else 
		   convert(REMOVE_WS(msg_sms) using euckr)
	     end) as msg_sms, 
		phn, 
		remark1, 
		remark2,
		result, 
		convert(REMOVE_WS(sms_lms_tit) using euckr) as sms_lms_tit, 
		sms_kind, 
		sms_sender, 
		res_dt, 
		reserve_dt, 
		(select ifnull(file1_path, "") from api_mms_images aa where aa.user_id = drr.userid and aa.mms_id = drr.mms_image_id) as mms_file1, 
		(select ifnull(file2_path, "") from api_mms_images aa where aa.user_id = drr.userid and aa.mms_id = drr.mms_image_id) as mms_file2, 
		(select ifnull(file3_path, "") from api_mms_images aa where aa.user_id = drr.userid and aa.mms_id = drr.mms_image_id) as mms_file3
		,(case when sms_kind = 'S' then length(convert(REMOVE_WS(msg_sms) using euckr)) else 100 end) as msg_len
		,userid
		,(select max(sms_len_check) from DHN_CLIENT_LIST dcl where dcl.user_id = drr.userid) as sms_len_check
	FROM DHN_RESULT drr 
	WHERE send_group = ?
	  and result = 'P'
      and userid = ?
	order by userid
	`
	resrows, err := db.Query(resquery, group_no, user_id)

	if err != nil {
		stdlog.Println("Result Table 조회 중 오류 발생")
		stdlog.Println(err)
		stdlog.Println(resquery)
	}
	defer resrows.Close()

	var smsBox SendReqTable
	var mmsBox SendReqTable
	var resBox = []SendResTable{}
	var apiErrBox = []string{}

	fcnt := 0
	smscnt := 0
	lmscnt := 0
	tcnt := 0
	reg, err := regexp.Compile("[^0-9]+")
	smsSeq := 1
	mmsSeq := 1

	kisaCode, converr := strconv.ParseUint(config.Conf.KISA_CODE, 10, 64)
	if converr != nil {
		kisaCode = 302190001
	}
	
	for resrows.Next() {
		resrows.Scan(&msgid, &code, &message, &message_type, &msg_sms, &phn, &remark1, &remark2, &result, &sms_lms_tit, &sms_kind, &sms_sender, &res_dt, &reserve_dt, &mms_file1, &mms_file2, &mms_file3, &msgLen, &userid, &sms_len_check)
		phnstr = phn.String

		// 알림톡 발송 성공 혹은 문자 발송이 아니면
		// API_RESULT 성공 처리 함.
		if len(msg_sms.String) > 0 && len(sms_sender.String) > 0 {
			phnstr = reg.ReplaceAllString(phnstr, "")
			if s.HasPrefix(phnstr, "82") {
				phnstr = "0" + phnstr[2:len(phnstr)]
			}
			
			if s.EqualFold(sms_kind.String, "S") {
				if msgLen.Int64 <= 90 || s.EqualFold(sms_len_check.String, "N") {
					smsBox = SendReqTable{
						MessageType : 1,
						MessageSubType : 1,
						CallbackNumber : s.ReplaceAll(sms_sender.String, "-", ""),
						CustomMessageID : msgid.String,
						KisaOrigCode : kisaCode,
						Bundle : []Bundle{
							{
								Seq : 1,
								Number : phnstr,
								Content : msg_sms.String,
							},
						},
					}

					resp, err := client.ExecSMS("/send/sms", smsBox)
					if err != nil {
						apiErrBox = append(apiErrBox, msgid.String)
						stdlog.Println(user_id, "- msgid : ", msgid.String, " KT크로샷 sms API 발송 중 오류 발생 : ", err)
						continue
					}
					if resp.StatusCode != 200 {
						apiErrBox = append(apiErrBox, msgid.String)
						stdlog.Println(user_id, "- msgid : ", msgid.String, " KT크로샷 sms API 발송 중 오류 발생 : ", err, "  /  statusCode : ", resp.StatusCode)
						continue
					}

					if smsSeq > 5 {
						smsSeq = 1
					}

					body, _ := ioutil.ReadAll(resp.Body)
					resBox = append(resBox, SendResTable{
						MsgID : msgid.String,
						SendReqTable : smsBox,
						MessageType : "sms",
						ResCode : resp.StatusCode,
						BodyData : body,
						Seq : smsSeq,
					})

					smsSeq++
					smscnt++
				} else {
					db.Exec("update DHN_RESULT dr set dr.result = 'Y', dr.code = '7003', dr.message = '메세지 길이 오류', dr.remark2 = date_format(now(), '%Y-%m-%d %H:%i:%S') where userid = '" + userid.String + "' and msgid = '" + msgid.String + "'")
				}
			} else if s.EqualFold(sms_kind.String, "L") || s.EqualFold(sms_kind.String, "M") {

				messageType := "lms"
				subType := 1
				var fileParam []string
				if mms_file1.String != "" {
					fileParam = append(fileParam, mms_file1.String)
					messageType = "mms"
					subType = 3
				} else {
					fileParam = append(fileParam, "")
				}
				if mms_file2.String != "" {
					fileParam = append(fileParam, mms_file2.String)
					messageType = "mms"
					subType = 3
				} else {
					fileParam = append(fileParam, "")
				}
				if mms_file3.String != "" {
					fileParam = append(fileParam, mms_file3.String)
					messageType = "mms"
					subType = 3
				} else {
					fileParam = append(fileParam, "")
				}

				mmsBox = SendReqTable{
					MessageType : 4,
					MessageSubType : subType,
					CallbackNumber : s.ReplaceAll(sms_sender.String, "-", ""),
					CustomMessageID : msgid.String,
					KisaOrigCode : kisaCode,
					Bundle : []Bundle{
						{
							Seq : 1,
							Number : phnstr,
							Content : msg_sms.String,
							Subject : sms_lms_tit.String,
						},
					},
				}
				

				resp, err := client.ExecMMS("/send/mms", mmsBox, fileParam)
				if err != nil {
					apiErrBox = append(apiErrBox, msgid.String)
					stdlog.Println(user_id, "- msgid : ", msgid.String, " KT크로샷 mms API 발송 중 오류 발생 : ", err)
					continue
				}
				if resp.StatusCode != 200 {
					apiErrBox = append(apiErrBox, msgid.String)
					stdlog.Println(user_id, "- msgid : ", msgid.String, " KT크로샷 sms API 발송 중 오류 발생 : ", err, "  /  statusCode : ", resp.StatusCode)
					continue
				}

				if mmsSeq > 5 {
					mmsSeq = 1
				}

				body, _ := ioutil.ReadAll(resp.Body)
				resBox = append(resBox, SendResTable{
					MsgID : msgid.String,
					SendReqTable : mmsBox,
					MessageType : messageType,
					FileParam : fileParam,
					ResCode : resp.StatusCode,
					BodyData : body,
					Seq : mmsSeq,
				})

				mmsSeq++
				lmscnt++
			}

		} else {
			errCode := ""
			if len(sms_sender.String) <= 0 {
				errCode = "7051"
			} else if len(msg_sms.String) <= 0 {
				errCode = "7011"
			} else {
				errCode = "7011"
			}
			db.Exec("update DHN_RESULT dr set dr.result = 'Y', dr.code='"+errCode+"', dr.message = concat(dr.message, ',문자 발송 정보 누락'),dr.remark2 = date_format(now(), '%Y-%m-%d %H:%i:%S') where userid = '" + userid.String + "' and msgid = '" + msgid.String + "'")
		}

	}
	if len(resBox) > 0 {
		tx, err := db.Begin()
		if err != nil {
			stdlog.Println(err)
		}
		stmtSMS, _ := tx.Prepare("insert into KT_SMS(userid, msgid, MessageSubType, CallbackNumber, Bundle_Seq, Bundle_Num, Bundle_Content, resp_JobID, resp_SubmitTime, sep_seq, dhn_id) values(?,?,?,?,?,?,?,?,?,?,?)")
		stmtMMS, _ := tx.Prepare("insert into KT_MMS(userid, msgid, MessageSubType, CallbackNumber, Bundle_Seq, Bundle_Num, Bundle_Content, Bundle_Subject, Image_path1, Image_path2, Image_path3, resp_JobID, resp_SubmitTime, sep_seq, dhn_id) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
		var decodeBody SendResDetileTable
		for _, val := range resBox {
			srt := val.SendReqTable
			json.Unmarshal([]byte(val.BodyData), &decodeBody)
			if val.MessageType == "sms" {
				_, err := stmtSMS.Exec(user_id, val.MsgID, srt.MessageSubType, srt.CallbackNumber, srt.Bundle[0].Seq, srt.Bundle[0].Number, srt.Bundle[0].Content, decodeBody.JobIDs[0].JobID, decodeBody.SubmitTime, val.Seq, acc)
				if err != nil {
					tx.Rollback()
					stdlog.Println(user_id, "- msgid : ", val.MsgID, " KT테이블 SMS insert 중 오류 발생 : ", err)
				}
			} else if val.MessageType == "lms" {
				_, err := stmtMMS.Exec(user_id, val.MsgID, srt.MessageSubType, srt.CallbackNumber, srt.Bundle[0].Seq, srt.Bundle[0].Number, srt.Bundle[0].Content, srt.Bundle[0].Subject, "", "", "", decodeBody.JobIDs[0].JobID, decodeBody.SubmitTime, val.Seq, acc)
				if err != nil {
					tx.Rollback()
					stdlog.Println(user_id, "- msgid : ", val.MsgID, " KT테이블 LMS insert 중 오류 발생 : ", err)
				}
			} else if val.MessageType == "mms" {
				_, err := stmtMMS.Exec(user_id, val.MsgID, srt.MessageSubType, srt.CallbackNumber, srt.Bundle[0].Seq, srt.Bundle[0].Number, srt.Bundle[0].Content, srt.Bundle[0].Subject, val.FileParam[0], val.FileParam[1], val.FileParam[2], decodeBody.JobIDs[0].JobID, decodeBody.SubmitTime, val.Seq, acc)
				if err != nil {
					tx.Rollback()
					stdlog.Println(user_id, "- msgid : ", val.MsgID, " KT테이블 MMS insert 중 오류 발생 : ", err)
				}
			}
		}
		err = tx.Commit()
		if err != nil {
			stdlog.Println(user_id, " KT테이블 insert commit 중 오류 발생 시작 : ", err)
			tx.Rollback()
			for _, val := range resBox {
				stdlog.Println(user_id, "- msgid : ", val.SendReqTable.CustomMessageID, " KT테이블 insert 중 오류 발생 : ", err)
			}
			stdlog.Println(user_id, " KT테이블 insert commit 중 오류 발생 끝 : ", err)
		}
	}
	if len(apiErrBox) > 0 {
		for _, id := range apiErrBox {
			db.Exec("update DHN_RESULT set send_group = null where msgid = ?", id)
			stdlog.Println(user_id, "- msgid : ", msgid.String, " KT크로샷 오류건 send_group null 처리")
		}
		fcnt++
		time.Sleep(5 * time.Second)

	}
	if smscnt > 0 || lmscnt > 0 || fcnt > 0 {
		stdlog.Println(user_id, "-", group_no, "문자 발송 처리 완료 ( ", tcnt, " ) : SMS -", smscnt, " , LMS -", lmscnt, ", 그룹넘버초기화 - ", fcnt, "  >> Process cnt : ", procCnt)
	}
	procCnt--
}
