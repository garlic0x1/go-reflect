# go-reflect
Renamed due to there being a Burp plugin named "reflector"  
A crawler that tests HTML forms for reflection  
Based on https://github.com/hakluke/hakrawler  

For every HTML form found while crawling, all input fields will be submitted with a hash to try to fit the type (email, text, password, etc), and hidden fields will be set to their default value.  If those hashes appear in a response you will be notified

Using the `-proxy` flag will disable TLS verification and allow traffic to be viewed in an intercept proxy

# Usage:
```
$ ./reflector -h
flag needs an argument: -h
Usage of ./reflector:
  -d int
    	Depth to crawl. (default 2)
  -h string
    	Custom headers separated by two semi-colons. E.g. -h "Cookie: foo=bar;;Referer: http://example.com/" 
  -insecure
    	Disable TLS verification.
  -proxy string
    	Proxy URL, example: -proxy http://127.0.0.1:8080
  -s	Show the source of URL based on where it was found (href, form, script, etc.)
  -subs
    	Include subdomains for crawling.
  -t int
    	Number of threads to utilise. (default 8)
  -u	Show only unique urls
```

# Example:
```
$ echo https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/ | ./reflector -u -s -d 3 -proxy http://localhost:8080
2022/03/11 21:32:24 Using proxy http://localhost:8080
[href] https://portswigger.net/web-security/cross-site-scripting/reflected/lab-html-context-nothing-encoded
[href] https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/
[href] https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/post?postId=3
[href] https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/post?postId=1
[href] https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/post?postId=2
[href] https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/post?postId=4
[href] https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/post?postId=5
[script] https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/resources/labheader/js/labHeader.js
[form] https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/
[form] https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/post/comment
[reflector] Injection from https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/ found at https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/?search=http%3A%2F%2FdEzNuRML
[reflector] Injection from https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/post/comment found at https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/post?postId=3
[href] http://yudsqgcx/
[href] http://nrufhnpa/
[reflector] Injection from https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/ found at https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/?search=http%3A%2F%2FaOrGBFPq
[reflector] Injection from https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/post/comment found at https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/post?postId=4
[href] http://hxhuvayu/
[reflector] Injection from https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/ found at https://ac7f1f701f2c6ea2c19f078f00eb00a7.web-security-academy.net/?search=http%3A%2F%2FvmdJFxaW

```
