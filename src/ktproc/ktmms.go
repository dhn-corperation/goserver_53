package ktproc

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	config "mycs/src/kaoconfig"
	databasepool "mycs/src/kaodatabasepool"
	"strconv"

	//	"log"
	s "strings"
	"sync"
	"time"

	"context"

	_ "github.com/go-sql-driver/mysql"
)

func LMSProcess(ctx context.Context, acc int) {
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():

			config.Stdlog.Println("Oshot MMS process가 20초 후에 종료 됨.")
			time.Sleep(20 * time.Second)
			config.Stdlog.Println("Oshot MMS process 종료 완료")
			return
		default:

			var t = time.Now()

			if t.Day() < 3 {
				for i := 1; i < 6; i++ {
					wg.Add(1)
					go mmsProcess(&wg, "KT_MMS", true, i, acc)
				}
			}

			for i := 1; i < 6; i++ {
				wg.Add(1)
				go mmsProcess(&wg, "KT_MMS", false, i, acc)
			}

			wg.Wait()
		}
	}

}

func mmsProcess(wg *sync.WaitGroup, table string, preFlag bool, seq int, acc int) {

	defer wg.Done()
	var db = databasepool.DB
	var errlog = config.Stdlog

	var isProc = true
	var t time.Time
	if preFlag {
		t = time.Now().Add(time.Hour * -96)
	} else {
		t = time.Now()
	}

	var monthStr = fmt.Sprintf("%d%02d", t.Year(), t.Month())

	var MMSTable = table + "_" + monthStr

	var tableQuery = "select 1 from " + MMSTable

	_, err := db.Query(tableQuery)
	if err != nil {
		errcode := err.Error()
		errlog.Println("KT크로샷 SMS LOG 테이블 존재유무 체크 ", tableQuery, errcode)

		if s.Index(errcode, "1146") > 0 {
			db.Exec("Create Table IF NOT EXISTS " + MMSTable + " like " + table)
			errlog.Println(MMSTable + " 생성 !!")
		}

		return
	}

	var searchQuery = "select userid, msgid, resp_JobID from " + table + " where sep_seq = " + strconv.Itoa(seq)

	searchData, err := db.Query(searchQuery)
	if err != nil {
		errcode := err.Error()
		errlog.Println("KT크로샷 MMS 조회 중 오류 발생", searchQuery, errcode)

		if s.Index(errcode, "1146") > 0 {
			db.Exec("Create Table IF NOT EXISTS " + MMSTable + " like " + table)
			errlog.Println(MMSTable + " 생성 !!")
		}

		isProc = false
		return
	}
	defer searchData.Close()

	if isProc {
		acc := account[acc]
		client := NewMessage(acc["apiKey"], acc["apiPw"], acc["userKey"], false, 3)
		for searchData.Next() {
			var userid, msgid sql.NullString
			var resp_JobID sql.NullInt64

			searchData.Scan(&userid, &msgid, &resp_JobID)

			sendData := SearchReqTable{
				JobIDs: []int64{
					resp_JobID.Int64,
				},
				SendDay: time.Now().Format("20060102"),
			}

			resp, err := client.SearchResult("/inquiry/report/", sendData)

			if err != nil {
				errlog.Println(userid.String, "- msgid : ", msgid.String, " KT크로샷 결과조회 API 발송 중 오류 발생 : err : ", err)
				continue
			}

			if resp.StatusCode != 200 {
				errlog.Println(userid.String, "- msgid : ", msgid.String, " KT크로샷 결과조회 API 발송 중 오류 발생 : statusCode : ", resp.StatusCode)
				continue
			}

			body, _ := ioutil.ReadAll(resp.Body)
			var decodeBody SearchResTable

			err = json.Unmarshal([]byte(body), &decodeBody)
			if err != nil {
				errlog.Println(userid.String, "- msgid : ", msgid.String, " KT크로샷 결과조 API 결과 변환중 에러 발생 : ", err)
				continue
			}
			first := decodeBody.JobIDs[0]

			testData, _ := json.Marshal(first)
			errlog.Println(string(testData))

			if first.Result == 0 {
				continue
			}

			convResult := strconv.Itoa(first.Result)
			resultCode := KTCode(convResult)
			resultMessage := KTCodeMessage(resultCode)

			errlog.Println("1")

			var telInfo = "ETC"
			var telInfoLog = 0
			if first.TelconInfo != nil {
				if *first.TelconInfo == 1 {
					telInfo = "SKT"
				} else if *first.TelconInfo == 2 {
					telInfo = "KTF"
				} else if *first.TelconInfo == 3 {
					telInfo = "LGT"
				}
				telInfoLog = *first.TelconInfo
			}
			errlog.Println("2")

			parsedTime, err := time.Parse("20060102150405", first.Time)
			if err != nil {
				errlog.Println(userid.String, "- msgid : ", msgid.String, " KT크로샷 결과조회 API 발송 중 시간변환 오류 발생 : ", err, "  /  statusCode : ", first.Time)
				continue
			}
			errlog.Println("3")

			formattedTime := parsedTime.Format("2006-01-02 15:04:05")

			_, err = db.Exec(`insert into ` + MMSTable + `(userid, msgid, MessageSubType, CallbackNumber, Bundle_Seq, Bundle_Num, Bundle_Content, Bundle_Subject, Image_path1, Image_path2, Image_path3, resp_JobID, resp_Time, resp_SubmitTime, resp_Result, Resp_TelconInfo, resp_EndUserID, resp_ServiceProviderID, sep_seq, dhn_id)
				select userid, msgid, MessageSubType, CallbackNumber, Bundle_Seq, Bundle_Num, Bundle_Content, Bundle_Subject, Image_path1, Image_path2, Image_path3, resp_JobID, '` + first.Time + `', '` + first.SubmitTime + `', '` + strconv.Itoa(first.Result) + `', '` + strconv.Itoa(telInfoLog) + `', '` + first.EndUserID + `', '` + first.ServiceProviderID + `', sep_seq, dhn_id
				from KT_MMS
				WHERE userid = '` + userid.String + `' and msgid = '` + msgid.String + `'`)
			if err != nil {
				errlog.Println(userid.String, "- msgid : ", msgid.String, " KT크로샷 결과조 API 결과 LOG 테이블 입력중 에러 발생 : ", err)
				continue
			}
			errlog.Println("4")
			if first.Result != 10000 {
				_, err = db.Exec("update DHN_RESULT set message_type = 'PH', result = 'Y', code = '" + resultCode + "', message = concat(message, '," + resultMessage + "'), remark1 = '" + telInfo + "', remark2 = '" + formattedTime + "' where userid='" + userid.String + "' and msgid = '" + msgid.String + "'")
				if err != nil {
					errlog.Println(userid.String, "- msgid : ", msgid.String, " KT크로샷 결과조 API 결과 DHN_RESULT 테이블 반영 실패1 : ", err)
				}
				errlog.Println("5")
			} else {
				_, err = db.Exec("update DHN_RESULT set message_type = 'PH', result = 'Y', code = '0000', message = '', remark1 = '" + telInfo + "', remark2 = '" + formattedTime + "' where userid='" + userid.String + "' and msgid = '" + msgid.String + "'")
				if err != nil {
					errlog.Println(userid.String, "- msgid : ", msgid.String, " KT크로샷 결과조 API 결과 DHN_RESULT 테이블 반영 실패2 : ", err)
				}
			}

			db.Exec(`delete from KT_MMS where userid = '` + userid.String + `' and msgid = '` + msgid.String + `'`)

		}
	}
}
