# reflector
A crawler that tests HTML forms for reflection  
Based on https://github.com/hakluke/hakrawler  

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

# Example output:
```
[href] https://garlic0x1.com/home?page=2
[form] https://garlic0x1.com/uploadfile
[form] https://garlic0x1.com/login
[form] https://garlic0x1.com/signup
[href] https://www.idontplaydarts.com/2012/06/encoding-web-shells-in-png-idat-chunks/
[href] https://github.com/garlic0x1/find_and_bypass_403
[href] https://medium.com/@insecurity_92477/utilizing-htaccess-for-exploitation-purposes-part-1-5733dd7fc8eb
[href] https://github.com/vavkamil/xss2png
[reflector] Injection from https://garlic0x1.com/signup found at https://garlic0x1.com/
[href] https://garlic0x1.com/logout
[href] https://garlic0x1.com/account
[reflector] Injection from https://garlic0x1.com/signup found at https://garlic0x1.com/login
```
