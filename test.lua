wrk.method = "POST"   -- POST 요청을 전송한다.
count      = 1
request    = function()
	count = count+1
	path = string.format("/message/2/%d", count)
	wrk.headers['x-application-name'] = "joinc" 
	return wrk.format(nil,path)
end
