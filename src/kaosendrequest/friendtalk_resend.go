package kaosendrequest

import(
	"fmt"
	"sync"
	"time"
	"context"
	"strconv"
	"database/sql"
	"encoding/json"
	s "strings"

	kakao "mycs/src/kakaojson"
	config "mycs/src/kaoconfig"
	databasepool "mycs/src/kaodatabasepool"
	cm "mycs/src/kaocommon"
	krt "mycs/src/kaoresulttable"
)



func FriendtalkResendProc(ctx context.Context) {
	procCnt := 0
	config.Stdlog.Println("Friendtalk 9999 resend - 프로세스 시작 됨 ")

	for {
		if procCnt < 3 {
			select {
			case <- ctx.Done():
			    config.Stdlog.Println("Friendtalk 9999 resend - process가 10초 후에 종료 됨.")
			    time.Sleep(10 * time.Second)
			    config.Stdlog.Println("Friendtalk 9999 resend - process 종료 완료")
			    return
			default:
				var count sql.NullInt64
				cnterr := databasepool.DB.QueryRowContext(ctx, "SELECT count(1) AS cnt FROM DHN_REQUEST_RESEND WHERE send_group IS NULL").Scan(&count)
				
				if cnterr != nil && cnterr != sql.ErrNoRows {
					config.Stdlog.Println("Friendtalk 9999 resend - DHN_REQUEST_RESEND Table - select 오류 : " + cnterr.Error())
					time.Sleep(10 * time.Second)
				} else {
					if count.Valid && count.Int64 > 0 {		
						var startNow = time.Now()
						var group_no = fmt.Sprintf("%02d%02d%02d%09d", startNow.Hour(), startNow.Minute(), startNow.Second(), startNow.Nanosecond())
						
						updateRows, err := databasepool.DB.ExecContext(ctx, "update DHN_REQUEST_RESEND set send_group = ? where send_group is null limit ?", group_no, strconv.Itoa(config.Conf.SENDLIMIT))
				
						if err != nil {
							config.Stdlog.Println("Friendtalk 9999 resend - send_group Update 오류 : ", err)
						}
				
						rowcnt, _ := updateRows.RowsAffected()
				
						if rowcnt > 0 {
							procCnt++
							config.Stdlog.Println("Friendtalk 9999 resend - 발송 처리 시작 ( ", group_no, " ) : ", rowcnt, " 건  ( Proc Cnt :", procCnt, ") - START")
							go func() {
								defer func() {
									procCnt--
								}()
								ftResendProcess(group_no, procCnt)
							}()
						}
					}
				}
			}
		}
	}
}

func ftResendProcess(group_no string, pc int) {
	defer func(){
		if r := recover(); r != nil {
			config.Stdlog.Println("Friendtalk 9999 resend - ftsendProcess panic 발생 원인 : ", r)
			if err, ok := r.(error); ok {
				if s.Contains(err.Error(), "connection refused") {
					for {
						config.Stdlog.Println("Friendtalk 9999 resend - ftsendProcess send ping to DB")
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

	ftColumn := cm.GetResFtColumn()
	ftColumnStr := s.Join(ftColumn, ",")

	var db = databasepool.DB
	var stdlog = config.Stdlog
	var errlog = config.Stdlog

	reqsql := "select * from DHN_REQUEST_RESEND where send_group = '" + group_no + "' and message_type like 'f%'"

	reqrows, err := db.Query(reqsql)
	if err != nil {
		errlog.Println("Friendtalk 9999 resend - ftsendProcess 쿼리 에러 group_no : ", group_no, " / query : ", reqsql)
		errlog.Println("Friendtalk 9999 resend - ftsendProcess 쿼리 에러 : ", err)
		panic(err)
	}
	defer reqrows.Close()

	columnTypes, err := reqrows.ColumnTypes()
	if err != nil {
		errlog.Println("Friendtalk 9999 resend - ftsendProcess 컬럼 초기화 에러 group_no : ", group_no)
		errlog.Println("Friendtalk 9999 resend - ftsendProcess 컬럼 초기화 에러 : ", err)
		time.Sleep(5 * time.Second)
	}
	count := len(columnTypes)
	initScanArgs := cm.InitDatabaseColumn(columnTypes, count)

	var procCount int
	procCount = 0
	var startNow = time.Now()
	var serial_number = fmt.Sprintf("%04d%02d%02d-", startNow.Year(), startNow.Month(), startNow.Day())

	resinsStrs := []string{}
	resinsValues := []interface{}{}
	resinsQuery := `insert IGNORE into DHN_RESULT(`+ftColumnStr+`) values %s`

	resultChan := make(chan krt.ResultStr, config.Conf.SENDLIMIT)
	var reswg sync.WaitGroup

	for reqrows.Next() {
		scanArgs := initScanArgs

		err := reqrows.Scan(scanArgs...)
		if err != nil {
			errlog.Println("Friendtalk 9999 resend - ftsendProcess column scan error : ", err, " / group_no : ", group_no)
			time.Sleep(5 * time.Second)
		}

		var friendtalk kakao.Friendtalk
		var attache kakao.Attachment
		var tcarousel kakao.TCarousel
		var carousel kakao.FCarousel
		var button []kakao.Button
		var image kakao.Image
		var coupon kakao.AttCoupon
		var itemList kakao.AttItem
		result := map[string]string{}

		for i, v := range columnTypes {

			switch s.ToLower(v.Name()) {
			case "msgid":
				if z, ok := (scanArgs[i]).(*sql.NullString); ok {
					friendtalk.Serial_number = serial_number + z.String
				}

			case "message_type":
				if z, ok := (scanArgs[i]).(*sql.NullString); ok {
					friendtalk.Message_type = s.ToUpper(z.String)
				}

			case "profile":
				if z, ok := (scanArgs[i]).(*sql.NullString); ok {
					friendtalk.Sender_key = z.String
				}

			case "phn":
				if z, ok := (scanArgs[i]).(*sql.NullString); ok {
					var cPhn string
					if s.HasPrefix(z.String, "0"){
						cPhn = s.Replace(z.String, "0", "82", 1)
					} else {
						cPhn = z.String
					}
					friendtalk.Phone_number = cPhn
				}

			case "msg":
				if z, ok := (scanArgs[i]).(*sql.NullString); ok {
					friendtalk.Message = z.String
				}

			case "ad_flag":
				if z, ok := (scanArgs[i]).(*sql.NullString); ok {
					friendtalk.Ad_flag = z.String
				}

			case "header":
				if z, ok := (scanArgs[i]).(*sql.NullString); ok {
					friendtalk.Header = z.String
				}

			case "carousel":
				if z, ok := (scanArgs[i]).(*sql.NullString); ok {
				    
					json.Unmarshal([]byte(z.String), &tcarousel)
					carousel.Tail = tcarousel.Tail
					  
					for ci, _ := range tcarousel.List {
						var catt kakao.CarouselAttachment
						var tcarlist kakao.CarouselList
						
						json.Unmarshal([]byte(tcarousel.List[ci].Attachment), &catt)
						
						tcarlist.Header = tcarousel.List[ci].Header
						tcarlist.Message = tcarousel.List[ci].Message
						tcarlist.Attachment = catt
						carousel.List = append(carousel.List, tcarlist)
					}
					if len(carousel.List) > 0 {
						friendtalk.Carousel = &carousel
					}  
				}

			case "image_url":
				if z, ok := (scanArgs[i]).(*sql.NullString); ok {
					image.Img_url = z.String
				}

			case "image_link":
				if z, ok := (scanArgs[i]).(*sql.NullString); ok {
					image.Img_link = z.String
				}

			case "att_items":
				if z, ok := (scanArgs[i]).(*sql.NullString); ok {
					error := json.Unmarshal([]byte(z.String), &itemList)
					if error == nil {
						attache.Item = &itemList
					}
				}

			case "att_coupon":
				if z, ok := (scanArgs[i]).(*sql.NullString); ok {
					error := json.Unmarshal([]byte(z.String), &coupon)
					if error == nil {
						attache.Coupon = &coupon
					}
				}

			case "button1":
				fallthrough
			case "button2":
				fallthrough
			case "button3":
				fallthrough
			case "button4":
				fallthrough
			case "button5":
				if z, ok := (scanArgs[i]).(*sql.NullString); ok {
					if len(z.String) > 0 {
						var btn kakao.Button

						json.Unmarshal([]byte(z.String), &btn)
						button = append(button, btn)
					}
				}
			}

			if z, ok := (scanArgs[i]).(*sql.NullString); ok {
				result[s.ToLower(v.Name())] = z.String
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt32); ok {
				result[s.ToLower(v.Name())] = string(z.Int32)
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt64); ok {
				result[s.ToLower(v.Name())] = string(z.Int64)
			}

		}

		if len(result["image_url"]) > 0 && s.EqualFold(result["message_type"], "FT") {
			friendtalk.Message_type = "FI"
			if s.EqualFold(result["wide"], "Y") {
				friendtalk.Message_type = "FW"
			}
		}

		attache.Buttons = button
		if len(image.Img_url) > 0 {
			attache.Ftimage = &image
		}
		friendtalk.Attachment = attache

		var temp krt.ResultStr
		temp.Result = result
		reswg.Add(1)
		go resendKakao(&reswg, resultChan, friendtalk, temp)

	}

	reswg.Wait()
	chanCnt := len(resultChan)

	ftQmarkStr := cm.GetQuestionMark(ftColumn)

	for i := 0; i < chanCnt; i++ {

		resChan := <-resultChan
		result := resChan.Result

		if resChan.Statuscode == 200 {

			var kakaoResp kakao.KakaoResponse
			json.Unmarshal(resChan.BodyData, &kakaoResp)

			var resdt = time.Now()
			var resdtstr = fmt.Sprintf("%4d-%02d-%02d %02d:%02d:%02d", resdt.Year(), resdt.Month(), resdt.Day(), resdt.Hour(), resdt.Minute(), resdt.Second())

			var resCode = kakaoResp.Code

			resinsStrs = append(resinsStrs, "("+ftQmarkStr+")")
			resinsValues = append(resinsValues, result["msgid"])
			resinsValues = append(resinsValues, result["userid"])
			resinsValues = append(resinsValues, result["ad_flag"])
			resinsValues = append(resinsValues, result["button1"])
			resinsValues = append(resinsValues, result["button2"])
			resinsValues = append(resinsValues, result["button3"])
			resinsValues = append(resinsValues, result["button4"])
			resinsValues = append(resinsValues, result["button5"])
			resinsValues = append(resinsValues, resCode) // 결과 code
			resinsValues = append(resinsValues, result["image_link"])
			resinsValues = append(resinsValues, result["image_url"])
			resinsValues = append(resinsValues, nil)               // kind
			resinsValues = append(resinsValues, kakaoResp.Message) // 결과 Message
			resinsValues = append(resinsValues, result["message_type"])
			resinsValues = append(resinsValues, result["msg"])
			resinsValues = append(resinsValues, result["msg_sms"])
			resinsValues = append(resinsValues, result["only_sms"])
			resinsValues = append(resinsValues, result["p_com"])
			resinsValues = append(resinsValues, result["p_invoice"])
			resinsValues = append(resinsValues, result["phn"])
			resinsValues = append(resinsValues, result["profile"])
			resinsValues = append(resinsValues, result["reg_dt"])
			resinsValues = append(resinsValues, result["remark1"])
			resinsValues = append(resinsValues, result["remark2"])
			resinsValues = append(resinsValues, 1)
			resinsValues = append(resinsValues, result["remark4"])
			resinsValues = append(resinsValues, result["remark5"])
			resinsValues = append(resinsValues, resdtstr) // res_dt
			resinsValues = append(resinsValues, result["reserve_dt"])

			if s.EqualFold(resCode,"0000") {
				resinsValues = append(resinsValues, "Y") // 
			} else if len(result["sms_kind"])>=1 {
				resinsValues = append(resinsValues, "P") // sms_kind 가 SMS / LMS / MMS 이면 문자 발송 시도
			} else {
				resinsValues = append(resinsValues, "Y") // 
			} 

			resinsValues = append(resinsValues, resCode)
			resinsValues = append(resinsValues, result["sms_kind"])
			resinsValues = append(resinsValues, result["sms_lms_tit"])
			resinsValues = append(resinsValues, result["sms_sender"])
			resinsValues = append(resinsValues, "N")
			resinsValues = append(resinsValues, result["tmpl_id"])
			resinsValues = append(resinsValues, result["wide"])
			resinsValues = append(resinsValues, nil) // send group
			resinsValues = append(resinsValues, result["supplement"])
			resinsValues = append(resinsValues, result["price"])
			resinsValues = append(resinsValues, result["currency_type"])
			resinsValues = append(resinsValues, result["header"])
			resinsValues = append(resinsValues, result["carousel"])
			resinsValues = append(resinsValues, result["mms_image_id"])
			resinsValues = append(resinsValues, result["att_items"])
			resinsValues = append(resinsValues, result["att_coupon"])

			if len(resinsStrs) >= 500 {
				resinsStrs, resinsValues = cm.InsMsg(resinsQuery, resinsStrs, resinsValues)
			}

		} else {
			stdlog.Println("Friendtalk 9999 resend - 친구톡 서버 처리 오류 : ( ", string(resChan.BodyData), " )", result["msgid"])
			db.Exec("update DHN_REQUEST_RESEND set send_group = null where msgid = '" + result["msgid"] + "'")
		}

		procCount++
	}

	if len(resinsStrs) > 0 {
		resinsStrs, resinsValues = cm.InsMsg(resinsQuery, resinsStrs, resinsValues)
	}

	stdlog.Println("Friendtalk 9999 resend - 발송 처리 완료 ( ", group_no, " ) : ", procCount, " 건  ( Proc Cnt :", pc, ") - END" )
	

}

func resendKakao(reswg *sync.WaitGroup, c chan<- krt.ResultStr, friendtalk kakao.Friendtalk, temp krt.ResultStr) {
	defer reswg.Done()

	var seq int = 1

	for {
		friendtalk.Serial_number = friendtalk.Serial_number + strconv.Itoa(seq)
		resp, err := config.Client.R().
			SetHeaders(map[string]string{"Content-Type": "application/json"}).
			SetBody(friendtalk).
			Post(config.Conf.API_SERVER + "v3/" + config.Conf.PROFILE_KEY + "/friendtalk/send")

		if err != nil {
			config.Stdlog.Println("Friendtalk 9999 resend - 친구톡 메시지 서버 호출 오류 : ", err)
		} else {
			temp.Statuscode = resp.StatusCode()
			if temp.Statuscode != 500 {
				temp.BodyData = resp.Body()
				break
			}
		}
		seq++
	}
	databasepool.DB.Exec("update DHN_REQUEST_RESEND set try_cnt = " + strconv.Itoa(seq) + " where msgid = '" + temp.Result["msgid"] + "'")
	c <- temp
}