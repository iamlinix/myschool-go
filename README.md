# An almost automatic MySchool crawler using [selenium](https://www.selenium.dev/)

- Search schools by state, automatically traverse each school in the list, scrape NAPLAN scores, including historical ones, and stores them in the db.

- The process maybe interrupted by recaptcha, the procedure will wait for upto 5 minutes for a human being to solve it and then carry on.

- Breakpoint supported, can stop/restart arbitarily.

- Only supports Mysql for now.

- **NOTICE: selenium does not have official golang binding, so it's not supported well with golang, the user will need to manually download the browser driver, like [Chrome driver](https://chromedriver.chromium.org/downloads) for instance**

- Build from source:
  - `go mod tidy`
  - `go build -o myschool .\cmd\myschool\myschool.go`
  - `.\myschool -c .\chromedriver.exe -s qld -o localhost -u myuser -p mypass -d mydb`
