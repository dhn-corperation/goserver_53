package kaoreqreceive

import (
	"database/sql"
	"time"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	config "mycs/src/kaoconfig"
	databasepool "mycs/src/kaodatabasepool"
	"mycs/src/kaoresulttable"

	"github.com/gin-gonic/gin"

	s "strings"
)


func SearchResultReq(c *gin.Context) {
	errlog := config.Stdlog
	db := databasepool.DB

	ctx := c.Request.Context()

	userid := c.Request.Header.Get("userid")
	userip := c.ClientIP()
	isValidation := false

	sqlstr := `
		select 
			count(1) as cnt 
		from
			DHN_CLIENT_LIST
		where
			user_id = ?
			and ip = ?
			and use_flag = 'Y'`

	var cnt int
	err := db.QueryRowContext(ctx, sqlstr, userid, userip).Scan(&cnt)
	if err != nil { errlog.Println(err) }

	if cnt > 0 { 
		isValidation = true 
	} else {
		errlog.Println("허용되지 않은 사용자 및 아이피에서 발송 요청!! (userid : ", userid, "/ ip : ", userip, ")")
	}

	var startNow = time.Now()
	var startTime = fmt.Sprintf("%02d:%02d:%02d", startNow.Hour(), startNow.Minute(), startNow.Second())

	if isValidation {
		db2json := map[string]string{
			"msgid":         "msgid",
			"userid":        "userid",
			"ad_flag":       "ad_flag",
			"button1":       "button1",
			"button2":       "button2",
			"button3":       "button3",
			"button4":       "button4",
			"button5":       "button5",
			"code":          "code",
			"image_link":    "image_link",
			"image_url":     "image_url",
			"kind":          "kind",
			"message":       "message",
			"message_type":  "message_type",
			"msg":           "msg",
			"msg_sms":       "msg_sms",
			"only_sms":      "only_sms",
			"p_com":         "p_com",
			"p_invoice":     "p_invoice",
			"phn":           "phn",
			"profile":       "profile",
			"reg_dt":        "reg_dt",
			"remark1":       "remark1",
			"remark2":       "remark2",
			"remark3":       "remark3",
			"remark4":       "remark4",
			"remark5":       "remark5",
			"res_dt":        "res_dt",
			"reserve_dt":    "reserve_dt",
			"result":        "result",
			"s_code":        "s_code",
			"sms_kind":      "sms_kind",
			"sms_lms_tit":   "sms_lms_tit",
			"sms_sender":    "sms_sender",
			"sync":          "sync",
			"tmpl_id":       "tmpl_id",
			"wide":          "wide",
			"send_group":    "send_group",
			"supplement":    "supplement",
			"price":         "price",
			"currency_type": "currency_type",
			"title":         "title",
		}
		
		var reqData kaoresulttable.ResultTable
		err1 := c.ShouldBindJSON(&reqData)

		if err1 != nil { errlog.Println(err1) }
		errlog.Println("발송 결과 재수신 시작 ( ", userid, ") : ", len(reqData.Msgid), startTime)

		if len(reqData.Msgid) == 0 {
			c.JSON(200, "msgid가 존재하지 않습니다.")
			return
		}

		for i, v := range reqData.Msgid {
			reqData.Msgid[i] = fmt.Sprintf("'%s'", v)
		}

		msgids := s.Join(reqData.Msgid, ", ")

		joinSql := ""
		joinTable := "DHN_RESULT_" + reqData.Regdt

		exists, err2 := checkTableExists(db, joinTable)
		if err2 != nil {
			errlog.Println("발송 결과 재수신 table 존재유무 조회 오류 err : ", err2)
		}

		if exists {
			joinSql = " union all select * from " + joinTable + " where userid = '" + userid + "' and msgid in (" + msgids + ")"
		} else {
			joinSql = " "
		}

		resultSql := "select * from DHN_RESULT_PROC where userid = '" + userid + "' and msgid in (" + msgids + ")" + joinSql

		reqrows, err := db.Query(resultSql)
		if err != nil {
			errlog.Fatal(resultSql, err)
		}

		columnTypes, err := reqrows.ColumnTypes()
		if err != nil {
			errlog.Fatal(err)
		}
		
		count := len(columnTypes)
		scanArgs := make([]interface{}, count)
		
		finalRows := []interface{}{}
		upmsgids := []interface{}{}

		var isContinue bool
		
		isFirstRow := true

		for reqrows.Next() {
			
			if isFirstRow {
				errlog.Println("결과 전송 ( ", userid, " ) : 시작 " )
				for i, v := range columnTypes {
	
					switch v.DatabaseTypeName() {
					case "VARCHAR", "TEXT", "UUID", "TIMESTAMP":
						scanArgs[i] = new(sql.NullString)
						break
					case "BOOL":
						scanArgs[i] = new(sql.NullBool)
						break
					case "INT4":
						scanArgs[i] = new(sql.NullInt64)
						break
					default:
						scanArgs[i] = new(sql.NullString)
					}
				}
				isFirstRow = false				
			}

			err := reqrows.Scan(scanArgs...)
			if err != nil {
				errlog.Fatal(err)
			}

			masterData := map[string]interface{}{}

			for i, v := range columnTypes {

				isContinue = false

				if z, ok := (scanArgs[i]).(*sql.NullBool); ok {
					masterData[db2json[s.ToLower(v.Name())]] = z.Bool
					isContinue = true
				}

				if z, ok := (scanArgs[i]).(*sql.NullString); ok {
					masterData[db2json[s.ToLower(v.Name())]] = z.String
					isContinue = true
				}

				if z, ok := (scanArgs[i]).(*sql.NullInt64); ok {
					masterData[db2json[s.ToLower(v.Name())]] = z.Int64
					isContinue = true
				}

				if z, ok := (scanArgs[i]).(*sql.NullFloat64); ok {
					masterData[db2json[s.ToLower(v.Name())]] = z.Float64
					isContinue = true
				}

				if z, ok := (scanArgs[i]).(*sql.NullInt32); ok {
					masterData[db2json[s.ToLower(v.Name())]] = z.Int32
					isContinue = true
				}
				if !isContinue {
					masterData[db2json[s.ToLower(v.Name())]] = scanArgs[i]
				}

				if s.EqualFold(v.Name(), "MSGID") {
					upmsgids = append(upmsgids, masterData[db2json[s.ToLower(v.Name())]])
				}
			}

			finalRows = append(finalRows, masterData)
		}

		

		if len(finalRows) > 0 {
			errlog.Println("결과 전송 ( ", userid, " ) : ", len(finalRows))
			
			var commastr = "update DHN_RESULT_PROC set sync='Y' where userid = '" + userid + "' and msgid in (?)"

			_, err := db.Exec(commastr, msgids)

			if err != nil {
				errlog.Println("searchResult Table Update 처리 중 오류 발생 ")
			}

		}
		c.JSON(200, finalRows)
	} else {
		c.JSON(404, gin.H{
			"code":    "error",
			"message": "사용자 아이디 확인",
			"userid":  userid,
			"ip":      userip,
		})
	}
}