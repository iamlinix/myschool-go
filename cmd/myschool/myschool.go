package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"myschool/internal/model"
	"myschool/internal/util"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/akamensky/argparse"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const MYSCHOOL_URL = "https://www.myschool.edu.au"

var (
	db      *gorm.DB
	browser selenium.WebDriver
)

func panicIfErr(err error) {
	if err == nil {
		return
	}
	panic(err)
}

func acceptTNC() {
	checkbox, _ := browser.FindElement(selenium.ByID, "checkBoxTou")
	acceptBtn, _ := browser.FindElement(selenium.ByID, "acceptButton")
	checkbox.Click()
	acceptBtn.Click()
}

func selectState(state string) {
	browser.SetImplicitWaitTimeout(5 * time.Second)
	browser.FindElement(selenium.ByID, "dropdown-2")
	browser.Get(MYSCHOOL_URL + "/school-search?FormPosted=True&SchoolSearchQuery=&SchoolSector=&SchoolType=&State=" + state)
}

func jump2Breakpoint() (int, int) {
	page, index := util.LoadBreakpoint()
	if page > 0 {
		currentURL, _ := browser.CurrentURL()
		browser.Get(fmt.Sprintf("%s&pagenumber=%d", currentURL, page+1))
	}

	return page, index
}

func getComparativeNumbers(element selenium.WebElement) (int, int, int, int, int, int, int) {
	base, genLow, genHigh, simLow, simHigh, simAvg, allAvg := 0, 0, 0, 0, 0, 0, 0

	baseText, err := element.Text()
	if err == nil {
		cut := strings.Index(baseText, "\n")
		if cut >= 0 {
			base, _ = strconv.Atoi(baseText[0:cut])
		}
	}

	generalDiv, err := element.FindElement(selenium.ByXPATH, "./span/table/tbody/tr[@class='selected-school-row']/td/span[@class='err']")
	if err == nil {
		generalAttr, _ := generalDiv.Text()
		parts := strings.Split(generalAttr, " - ")
		if len(parts) == 2 {
			genLow, _ = strconv.Atoi(parts[0])
			genHigh, _ = strconv.Atoi(parts[1])
		}
	}

	cols, err := element.FindElements(selenium.ByXPATH, "./span/table/tbody/tr[@class='sim-all-row'][1]/td")
	if err == nil && len(cols) >= 2 {
		allDiv, err := cols[1].FindElement(selenium.ByXPATH, "./span[@class='sim-avg']")
		if err == nil {
			allAttr, err := allDiv.Text()
			if err == nil {
				allAvg, _ = strconv.Atoi(allAttr)
			}
		}

		spans, _ := cols[0].FindElements(selenium.ByTagName, "span")
		for _, span := range spans {
			cls, _ := span.GetAttribute("class")
			text, _ := span.Text()

			if cls == "sim-avg" {
				simAvg, _ = strconv.Atoi(text)
			} else if cls == "err" {
				sims := strings.Split(text, " - ")
				if len(sims) > 0 {
					simLow, _ = strconv.Atoi(sims[0])
				}
				if len(sims) > 1 {
					simHigh, _ = strconv.Atoi(sims[1])
				}
			}
		}
	}

	return base, genLow, genHigh, simLow, simHigh, simAvg, allAvg
}

func traverseSchools(page, index int) {
	link_index := index
	for {
		browser.SetImplicitWaitTimeout(300 * time.Second)
		schools, _ := browser.FindElements(selenium.ByCSSSelector, ".school-section")
		var schoolLinks []string

		for _, school := range schools {
			aTag, _ := school.FindElement(selenium.ByTagName, "a")
			href, _ := aTag.GetAttribute("href")
			schoolLinks = append(schoolLinks, href)
		}

		if link_index > 0 {
			schoolLinks = schoolLinks[link_index:]
		}

		for _, link := range schoolLinks {
			browser.DeleteAllCookies()
			browser.ExecuteScript("window.open()", nil)
			hanldes, _ := browser.WindowHandles()
			browser.SwitchWindow(hanldes[len(hanldes)-1])

			browser.SetImplicitWaitTimeout(300 * time.Second)
			browser.Get(MYSCHOOL_URL + link)
			acceptTNC()

			topSection, _ := browser.FindElement(selenium.ByCSSSelector, ".topsection-wrapper")
			headerSection, _ := topSection.FindElement(selenium.ByTagName, "h1")
			headerText, _ := headerSection.Text()
			headers := strings.Split(headerText, ",")

			schoolName := strings.Trim(headers[0], " ")
			suburb := strings.Trim(headers[1], " ")
			state := strings.Trim(headers[len(headers)-1], " ")

			t := time.Now()
			var factDiv selenium.WebElement
			var err error
			for {
				factDiv, err = browser.FindElement(selenium.ByCSSSelector, ".school-facts")
				if err != nil {
					time.Sleep(500 * time.Millisecond)
					if time.Since(t).Seconds() > 300 {
						panic(err)
					}
					continue
				}

				break
			}

			facts, _ := factDiv.FindElements(selenium.ByXPATH, "./ul/li")
			schoolSectorEle, _ := facts[0].FindElement(selenium.ByXPATH, "./div[@class='col2']")
			schoolSector, _ := schoolSectorEle.Text()
			schoolTypeEle, _ := facts[1].FindElement(selenium.ByXPATH, "./div[@class='col2']")
			schoolType, _ := schoolTypeEle.Text()
			yearRangeEle, _ := facts[2].FindElement(selenium.ByXPATH, "./div[@class='col2']")
			yearRange, _ := yearRangeEle.Text()
			schoolLocationEle, _ := facts[3].FindElement(selenium.ByXPATH, "./div[@class='col2']")
			schoolLocation, _ := schoolLocationEle.Text()

			url, _ := browser.CurrentURL()
			browser.Get(url + "/naplan/results")

			browser.SetImplicitWaitTimeout(time.Second)
			owls, _ := browser.FindElements(selenium.ByCSSSelector, ".owl-item")
			var owlLinks []string
			for _, owl := range owls {
				owlEle, err := owl.FindElement(selenium.ByTagName, "a")
				if err != nil {
					continue
				}

				link, err := owlEle.GetAttribute("href")
				if err != nil || len(link) == 0 {
					continue
				}

				owlLinks = append(owlLinks, link)
			}

			for _, link := range owlLinks {
				browser.Get(MYSCHOOL_URL + link)
				table, _ := browser.FindElement(selenium.ByID, "similarSchoolsTable")
				rows, _ := table.FindElements(selenium.ByXPATH, "./tbody/tr")
				linkParts := strings.Split(link, "/")
				year, _ := strconv.Atoi(linkParts[len(linkParts)-1])

				browser.ExecuteScript("[].slice.call(document.getElementsByClassName('popup-tooltiptext')).map(x => x.style.display = 'block')", nil)

				var scores []model.Score
				for _, row := range rows {
					cols, _ := row.FindElements(selenium.ByXPATH, "./td")
					gradeText, _ := cols[0].Text()
					grade := strings.Trim(gradeText, " ")
					reading, rGenLow, rGenHigh, rSimLow, rSimHigh, rSimAvg, rAllAvg := getComparativeNumbers(cols[1])
					writing, wGenLow, wGenHigh, wSimLow, wSimHigh, wSimAvg, wAllAvg := getComparativeNumbers(cols[2])
					spelling, sGenLow, sGenHigh, sSimLow, sSimHigh, sSimAvg, sAllAvg := getComparativeNumbers(cols[3])
					grammar, gGenLow, gGenHigh, gSimLow, gSimHigh, gSimAvg, gAllAvg := getComparativeNumbers(cols[4])
					numeracy, nGenLow, nGenHigh, nSimLow, nSimHigh, nSimAvg, nAllAvg := getComparativeNumbers(cols[5])
					scores = append(scores, model.Score{
						SchoolName:      schoolName,
						SchoolSector:    schoolSector,
						SchoolType:      schoolType,
						SchoolLocation:  schoolLocation,
						Suburb:          suburb,
						State:           state,
						Year:            year,
						Grade:           grade,
						YearRange:       yearRange,
						Reading:         reading,
						ReadingLow:      rGenLow,
						ReadingHigh:     rGenHigh,
						ReadingSimLow:   rSimLow,
						ReadingSimHigh:  rSimHigh,
						ReadingSimAvg:   rSimAvg,
						ReadingAllAvg:   rAllAvg,
						Writing:         writing,
						WritingLow:      wGenLow,
						WritingHigh:     wGenHigh,
						WritingSimLow:   wSimLow,
						WritingSimHigh:  wSimHigh,
						WritingSimAvg:   wSimAvg,
						WritingAllAvg:   wAllAvg,
						Spelling:        spelling,
						SpellingLow:     sGenLow,
						SpellingHigh:    sGenHigh,
						SpellingSimLow:  sSimLow,
						SpellingSimHigh: sSimHigh,
						SpellingSimAvg:  sSimAvg,
						SpellingAllAvg:  sAllAvg,
						Grammar:         grammar,
						GrammarLow:      gGenLow,
						GrammarHigh:     gGenHigh,
						GrammarSimLow:   gSimLow,
						GrammarSimHigh:  gSimHigh,
						GrammarSimAvg:   gSimAvg,
						GrammarAllAvg:   gAllAvg,
						Numeracy:        numeracy,
						NumeracyLow:     nGenLow,
						NumeracyHigh:    nGenHigh,
						NumeracySimLow:  nSimLow,
						NumeracySimHigh: nSimHigh,
						NumeracySimAvg:  nSimAvg,
						NumeracyAllAvg:  nAllAvg,
						Total:           reading + writing + spelling + grammar + numeracy,
					})
				}

				db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(&scores, 10)
			}
			util.SaveBreakpoint(page, link_index)

			browser.Close()
			handles, _ := browser.WindowHandles()
			browser.SwitchWindow(handles[0])
			link_index += 1

			x, _ := rand.Int(rand.Reader, big.NewInt(5))
			time.Sleep(time.Duration(x.Int64()) * time.Second)
		}

		pages, _ := browser.FindElement(selenium.ByCSSSelector, ".pagination")
		aTags, _ := pages.FindElements(selenium.ByTagName, "a")
		if len(aTags) == 0 {
			break
		}

		nextPage := aTags[len(aTags)-1]
		arrows, err := nextPage.FindElements(selenium.ByXPATH, "./i[@class='pag_arrow_right']")
		if len(arrows) > 0 {
			href, _ := nextPage.GetAttribute("href")
			browser.SetImplicitWaitTimeout(300 * time.Second)
			browser.Get(MYSCHOOL_URL + href)
			browser.FindElement(selenium.ByCSSSelector, ".showing-results")
		} else {
			panic(err)
		}
		page += 1
		link_index = 0
	}
}

func main() {
	parser := argparse.NewParser("myschool", "scrape NAPLAN scores from myschool.com.au")
	pstate := parser.String("s", "state", &argparse.Options{Required: false, Default: "QLD", Help: "The state to crawl"})
	pchrome := parser.String("c", "chrome", &argparse.Options{Required: false, Default: "chromedriver.exe", Help: "Chrome driver path"})
	puser := parser.String("u", "user", &argparse.Options{Required: true, Help: "Database username"})
	ppass := parser.String("p", "pass", &argparse.Options{Required: true, Help: "Database password"})
	phost := parser.String("o", "host", &argparse.Options{Required: false, Default: "localhost", Help: "Database server address"})
	pdb := parser.String("d", "database", &argparse.Options{Required: false, Default: "score", Help: "Default database"})
	pheadless := parser.Flag("l", "headless", &argparse.Options{Required: false, Help: "Run Chrome in headless mode"})
	err := parser.Parse(os.Args)
	panicIfErr(err)

	db, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True", *puser, *ppass, *phost, *pdb)), &gorm.Config{})
	panicIfErr(err)

	db.AutoMigrate(&model.Score{})

	service, err := selenium.NewChromeDriverService(*pchrome, 4444)
	panicIfErr(err)
	defer service.Stop()

	caps := selenium.Capabilities{"browserName": "chrome"}
	if *pheadless {
		caps.AddChrome(chrome.Capabilities{
			Args: []string{
				"--headless",
				"--ignore-certificate-errors",
				"--no-sandbox",
				"--disable-gpu",
			},
			W3C: true,
		})
	}

	browser, err = selenium.NewRemote(caps, "")
	panicIfErr(err)
	defer browser.Quit()

	browser.Get(MYSCHOOL_URL)
	browser.MaximizeWindow("")

	acceptTNC()
	selectState(*pstate)
	traverseSchools(jump2Breakpoint())
}
