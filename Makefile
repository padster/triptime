run:
	goapp serve triptime

deploy:
	appcfg.py -A triptime-1330 -V v1 update triptime/

test:
	curl -X POST -d @data/postdata.txt http://localhost:8080/_/verify

testtext:
	curl -X POST -d @data/posttext.txt http://localhost:8080/_/verify

testpostback:
	curl -X POST -d @data/postback.txt http://localhost:8080/_/verify

