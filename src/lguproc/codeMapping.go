package lguproc

func CodeMessage(code string) string {
	errmsg := map[string]string{
		"7000": "초기 입력 상태 (default)",
		"7001": "전송 요청 완료(결과수신대기)",
		"7003": "메시지 형식 오류",
		"7005": "휴대폰번호 가입자 없음(미등록)",
		"7006": "전송 성공",
		"7007": "결번(or 서비스 정지)",
		"7008": "단말기 전원 꺼짐",
		"7009": "단말기 음영지역",
		"7010": "단말기내 수신메시지함 FULL로 전송 실패 (구:단말 Busy, 기타 단말문제)",
		"7011": "기타 전송실패",
		"7013": "스팸차단 발신번호",
		"7014": "스팸차단 수신번호",
		"7015": "스팸차단 메시지내용",
		"7016": "스팸차단 기타",
		"7020": "*단말기 서비스 불가",
		"7021": "단말기 서비스 일시정지",
		"7022": "단말기 착신 거절",
		"7023": "단말기 무응답 및 통화중 (busy)",
		"7028": "단말기 MMS 미지원",
		"7029": "기타 단말기 문제",
		"7036": "유효하지 않은 수신번호(망)",
		"7037": "유효하지 않은 발신번호(망)",
		"7050": "이통사 컨텐츠 에러",
		"7051": "이통사 전화번호 세칙 미준수 발신번호",
		"7052": "이통사 발신번호 변작으로 등록된 발신번호",
		"7053": "이통사 번호도용문자 차단서비스에 가입된 발신번호",
		"7054": "이통사 발신번호 기타",
		"7059": "이통사 기타",
		"7060": "컨텐츠 크기 오류(초과 등)",
		"7061": "잘못된 메시지 타입",
		"7069": "컨텐츠 기타",
		"7074": "[Agent] 중복발송 차단 (동일한 수신번호와 메시지 발송 - 기본off, 설정필요)",
		"7075": "[Agent] 발송 Timeout",
		"7076": "[Agent] 유효하지않은 발신번호",
		"7077": "[Agent] 유효하지않은 수신번호",
		"7078": "[Agent] 컨텐츠 오류 (MMS파일없음 등)",
		"7079": "[Agent] 기타",
		"7080": "고객필터링 차단 (발신번호, 수신번호, 메시지 등)",
		"7081": "080 수신거부",
		"7084": "중복발송 차단",
		"7086": "유효하지 않은 수신번호",
		"7087": "유효하지 않은 발신번호",
		"7088": "발신번호 미등록 차단",
		"7089": "시스템필터링 기타",
		"7090": "발송제한 시간 초과",
		"7092": "잔액부족",
		"7093": "월 발송량 초과",
		"7094": "일 발송량 초과",
		"7095": "초당 발송량 초과 (재전송 필요)",
		"7096": "발송시스템 일시적인 부하 (재전송 필요)",
		"7097": "전송 네트워크 오류 (재전송 필요)",
		"7098": "외부발송시스템 장애 (재전송 필요)",
		"7099": "발송시스템 장애 (재전송 필요)",
	}
	val, ex := errmsg[code]
	if !ex {
		val = "기타 오류"
	}
	return val 
}

func LguCode(code string) string {
	mapTable := map[string]string{
		"01":    "7059",
		"5":     "7099",
		"7":     "7098",
		"11":    "7003",
		"12":    "7079",
		"13":    "7079",
		"14":    "7079",
		"15":    "7079",
		"16":    "7079",
		"17":    "7092",
		"18":    "7097",
		"21":    "7094",
		"22":    "7093",
		"25":    "7095",
		"30":    "7079",
		"31":    "7005",
		"32":    "7003",
		"33":    "7003",
		"34":    "7081",
		"50":    "7069",
		"51":    "7060",
		"52":    "7003",
		"55":    "7011",
		"56":    "7078",
		"100":   "7076",
		"101":   "7088",
		"102":   "7013",
		"103":   "7014",
		"110":   "7011",
		"111":   "7090",
		"200":   "7011",
		"201":   "7059",
		"202":   "7059",
		"250":   "7095",
		"500":   "7079",
		"501":   "7074",
		"502":   "7075",
		"503":   "7078",
		"504":   "7078",
		"505":   "7078",
		"506":   "7079",
		"510":   "7079",
		"1013":  "7083",
		"1014":  "7003",
		"1019":  "7098",
		"1022":  "7011",
		"1023":  "7059",
		"1024":  "7010",
		"1025":  "7008",
		"1026":  "7009",
		"1027":  "7022",
		"1999":  "7011",
		"2007":  "7011",
		"2008":  "7011",
		"2009":  "7011",
		"2010":  "7011",
		"2011":  "7011",
		"2012":  "7011",
		"3011":  "7022",
		"3012":  "7015",
		"3018":  "7013",
		"3019":  "7015",
		"3020":  "7074",
		"3030":  "7029",
		"3040":  "7029",
		"7000":  "7014",
		"8001":  "7061",
		"8004":  "7061",
		"8005":  "7011",
		"8006":  "7028",
		"10001": "7011",
		"10002": "7011",
		"10003": "7011",
		"10004": "7011",
		"10005": "7011",
		"10006": "7011",
		"10007": "7011",
		"10008": "7011",
		"10009": "7011",
		"10010": "7011",
		"10011": "7011",
		"10012": "7011",
		"10601": "7011",
		"10602": "7011",
		"10603": "7011",
		"10888": "7011",
		"10910": "7011",
		"10999": "7011",
		"11001": "7011",
		"11002": "7011",
		"11003": "7011",
		"11004": "7011",
		"11005": "7011",
		"11006": "7011",
		"11007": "7011",
		"11008": "7011",
		"11009": "7011",
		"11010": "7011",
		"11011": "7011",
		"11012": "7011",
		"11021": "7011",
		"11022": "7011",
		"11023": "7011",
		"11024": "7011",
		"11025": "7011",
		"11030": "7011",
		"11031": "7011",
		"11032": "7011",
		"11307": "7011",
		"12003": "7011",
		"12004": "7011",
		"13000": "7011",
		"13005": "7011",
		"13006": "7011",
		"13008": "7011",
		"13010": "7011",
		"13011": "7011",
		"13012": "7011",
		"13013": "7011",
		"13014": "7011",
		"13015": "7011",
		"13016": "7011",
		"13018": "7011",
		"13019": "7011",
		"13020": "7011",
		"13022": "7011",
		"13023": "7011",
		"13024": "7011",
		"14000": "7011",
		"14001": "7011",
		"15001": "7011",
		"15002": "7011",
		"15003": "7011",
		"15004": "7011",
		"15005": "7011",
		"15006": "7011",
		"15007": "7011",
		"15008": "7011",
		"15009": "7011",
		"15010": "7011",
		"15011": "7011",
		"15012": "7011",
		"15013": "7011",
		"15014": "7011",
		"15015": "7011",
		"15016": "7011",
		"15017": "7011",
		"15018": "7011",
		"15019": "7011",
		"15020": "7011",
		"15021": "7011",
		"15022": "7011",
		"15023": "7011",
		"15024": "7011",
		"15025": "7011",
		"15026": "7011",
		"15027": "7011",
		"15028": "7011",
		"15029": "7011",
		"15030": "7011",
		"15031": "7011",
		"15032": "7011",
		"15033": "7011",
		"19998": "7011",
		"19999": "7011",
		"20100": "7011",
		"20101": "7011",
		"20102": "7011",
		"20103": "7011",
		"20104": "7011",
		"20105": "7011",
		"20106": "7011",
		"20107": "7011",
		"20109": "7011",
		"20110": "7011",
		"20111": "7011",
		"20112": "7011",
		"20113": "7011",
		"20114": "7011",
		"20200": "7011",
		"20999": "7011",
		"30011": "7011",
		"30012": "7011",
		"30013": "7011",
		"30014": "7011",
		"30091": "7011",
		"30092": "7011",
		"30093": "7011",
		"30094": "7011",
		"30095": "7011",
		"30096": "7011",
		"30097": "7011",
		"30098": "7011",
		"30099": "7011",
		"30100": "7011",
		"30101": "7011",
		"30102": "7011",
		"30103": "7011",
		"30104": "7011",
		"30105": "7011",
		"30106": "7011",
		"30107": "7011",
		"30108": "7011",
		"30109": "7011",
		"30110": "7011",
		"30111": "7011",
		"30112": "7011",
		"30113": "7011",
		"30114": "7011",
		"30115": "7011",
		"30121": "7011",
		"30122": "7011",
		"30123": "7011",
		"30124": "7011",
		"30125": "7011",
		"30126": "7011",
		"30127": "7011",
		"30128": "7011",
		"30129": "7011",
		"30130": "7011",
		"30199": "7011",
		"30510": "7011",
		"30999": "7011",
		"60001": "7011",
		"60002": "7011",
		"60003": "7011",
		"60004": "7011",
		"60005": "7011",
		"60006": "7011",
		"60007": "7011",
		"60008": "7011",
		"60009": "7011",
		"60010": "7011",
		"60011": "7011",
		"60012": "7011",
		"60013": "7011",
		"60014": "7011",
		"60015": "7011",
		"60016": "7011",
		"60017": "7011",
		"60018": "7011",
		"60019": "7011",
		"60020": "7011",
		"60021": "7011",
		"60022": "7011",
		"60023": "7011",
		"60024": "7011",
		"60025": "7011",
		"60026": "7011",
		"60027": "7011",
		"60028": "7011",
		"60029": "7011",
		"60030": "7011",
		"60031": "7011",
		"60032": "7011",
		"60033": "7011",
		"60034": "7011",
		"60035": "7011",
		"60036": "7011",
		"60037": "7011",
		"60038": "7011",
		"60039": "7011",
		"60040": "7011",
		"60041": "7011",
		"60042": "7011",
		"60043": "7011",
		"60044": "7011",
		"60045": "7011",
		"60046": "7011",
		"60047": "7011",
		"60048": "7011",
		"60049": "7011",
		"60050": "7011",
		"60051": "7011",
		"60052": "7011",
		"60053": "7011",
		"60054": "7011",
		"60055": "7011",
		"60056": "7011",
		"60057": "7011",
		"60058": "7011",
		"60059": "7011",
		"60060": "7011",
		"60061": "7011",
		"60062": "7011",
		"60063": "7011",
		"60064": "7011",
		"60065": "7011",
		"60066": "7011",
		"60067": "7011",
		"60068": "7011",
		"60069": "7011",
		"60070": "7011",
		"60071": "7011",
		"60072": "7011",
		"60073": "7011",
		"60074": "7011",
		"60075": "7011",
		"60076": "7011",
		"60077": "7011",
		"60078": "7011",
		"60079": "7011",
		"60080": "7011",
		"60081": "7011",
		"60082": "7011",
		"60083": "7011",
		"60084": "7011",
		"60085": "7011",
		"60086": "7011",
		"60087": "7011",
		"60088": "7011",
		"60089": "7011",
		"60090": "7011",
		"60091": "7011",
		"60092": "7011",
		"60093": "7011",
		"60094": "7011",
		"60095": "7011",
		"60096": "7011",
		"60097": "7011",
		"60098": "7011",
		"60099": "7011",
		"60100": "7011",
		"60101": "7011",
		"60102": "7011",
		"60103": "7011",
		"60104": "7011",
		"60105": "7011",
		"60106": "7011",
		"60107": "7011",
		"60108": "7011",
		"60109": "7011",
		"60110": "7011",
		"60111": "7011",
		"60112": "7011",
		"60113": "7011",
		"60114": "7011",
		"60115": "7011",
		"60116": "7011",
		"60117": "7011",
		"60118": "7011",
		"60119": "7011",
		"60120": "7011",
		"60121": "7011",
		"60122": "7011",
		"60123": "7011",
		"60124": "7011",
		"60125": "7011",
		"70001": "7011",
		"71000": "7011",
		"71001": "7011",
		"71002": "7011",
		"71003": "7011",
		"71004": "7011",
		"71005": "7011",
		"71006": "7011",
		"71007": "7011",
		"71008": "7011",
		"71009": "7011",
		"71010": "7011",
		"71011": "7011",
		"71012": "7011",
		"71013": "7011",
		"71014": "7011",
		"71015": "7011",
		"71016": "7011",
		"71017": "7011",
		"71018": "7011",
		"71019": "7011",
		"71020": "7011",
		"71021": "7011",
		"71022": "7011",
		"71023": "7011",
		"71024": "7011",
		"71025": "7011",
		"71026": "7011",
		"71027": "7011",
		"71028": "7011",
		"71029": "7011",
		"71030": "7011",
		"71031": "7011",
		"71032": "7011",
		"71033": "7011",
		"71034": "7011",
		"71035": "7011",
		"71036": "7011",
		"71037": "7011",
		"71038": "7011",
		"71039": "7011",
		"71040": "7011",
		"71041": "7011",
		"71042": "7011",
		"71043": "7011",
		"71044": "7011",
		"71045": "7011",
		"71046": "7011",
		"71047": "7011",
		"71048": "7011",
		"71049": "7011",
		"71050": "7011",
		"71051": "7011",
		"71052": "7011",
		"71053": "7011",
		"71054": "7011",
		"71055": "7011",
		"71056": "7011",
		"71057": "7011",
		"71058": "7011",
		"71059": "7011",
		"71060": "7011",
		"71061": "7011",
		"71062": "7011",
		"71063": "7011",
		"71064": "7011",
		"71065": "7011",
		"71066": "7011",
		"71067": "7011",
		"71068": "7011",
		"71069": "7011",
		"71070": "7011",
		"71071": "7011",
		"71072": "7011",
		"71073": "7011",
		"71074": "7011",
		"71075": "7011",
		"71076": "7011",
		"71077": "7011",
		"71078": "7011",
		"71079": "7011",
		"71080": "7011",
		"71081": "7011",
		"71082": "7011",
		"71083": "7011",
		"71084": "7011",
		"71085": "7011",
		"71086": "7011",
		"71087": "7011",
		"71088": "7011",
		"71089": "7011",
		"71090": "7011",
		"71091": "7011",
		"71092": "7011",
		"71093": "7011",
		"71094": "7011",
		"71095": "7011",
		"80010": "7011",
		"80012": "7011",
		"80021": "7011",
		"80060": "7011",
		"80061": "7011",
		"80062": "7011",
		"80063": "7011",
		"80071": "7011",
		"80081": "7011",
		"80082": "7011",
		"80083": "7011",
		"80085": "7011",
		"80086": "7011",
		"80089": "7011",
		"80095": "7011",
		"80098": "7011",
		"80099": "7011",
		"80101": "7011",
		"80102": "7011",
		"80103": "7011",
		"80104": "7011",
		"80105": "7011",
		"80107": "7011",
		"80108": "7011",
		"80109": "7011",
		"80110": "7011",
		"80111": "7011",
		"80113": "7011",
		"80114": "7011",
		"80117": "7011",
		"80118": "7011",
		"80119": "7011",
		"80120": "7011",
		"80121": "7011",
		"80122": "7011",
		"80123": "7011",
		"80124": "7011",
		"80125": "7011",
		"80126": "7011",
		"80127": "7011",
		"80140": "7011",
		"80170": "7011",
		"80180": "7011",
		"80181": "7011",
		"80182": "7011",
		"80183": "7011",
		"80184": "7011",
		"80185": "7011",
		"81100": "7011",
		"82000": "7011",
		"82001": "7011",
		"82002": "7011",
		"82003": "7011",
		"82004": "7011",
		"82005": "7011",
		"82006": "7011",
		"82007": "7011",
		"83000": "7011",
		"83001": "7011",
		"83002": "7011",
		"83003": "7011",
		"83004": "7011",
		"83005": "7011",
		"83006": "7011",
		"84000": "7011",
		"84001": "7011",
		"84002": "7011",
		"84003": "7011",
		"85000": "7011",
		"89001": "7011",
		"89002": "7011",
		"89003": "7011",
		"89004": "7011",
		"89005": "7011",
		"89006": "7011",
		"89007": "7011",
	}
	val, ex := mapTable[code]
	if !ex {
		val = "7011"
	}
	return val
}
