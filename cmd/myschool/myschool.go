package main

import (
	"fmt"
	"myschool/internal/model"
	"myschool/internal/util"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/akamensky/argparse"
	"github.com/tebeka/selenium"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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

func main() {
	parser := argparse.NewParser("myschool", "scrape NAPLAN scores from myschool.com.au")
	pstate := parser.String("s", "state", &argparse.Options{Required: false, Default: "QLD", Help: "The state to crawl"})
	pchrome := parser.String("c", "chrome", &argparse.Options{Required: false, Default: "C:\\Users\\iamli\\source\\repos\\myschool-go\\chromedriver.exe", Help: "Chrome driver path"})
	err := parser.Parse(os.Args)
	panicIfErr(err)

	db, err = gorm.Open(mysql.Open("linx:qwer1234@tcp(localhost)/score?charset=utf8mb4&parseTime=True"), &gorm.Config{})
	panicIfErr(err)

	db.AutoMigrate(&model.Score{})

	service, err := selenium.NewChromeDriverService(*pchrome, 4444)
	panicIfErr(err)
	defer service.Stop()

	caps := selenium.Capabilities{"browserName": "chrome"}
	browser, err = selenium.NewRemote(caps, "")
	panicIfErr(err)
	defer browser.Quit()

	browser.Get(MYSCHOOL_URL)
	browser.MaximizeWindow("")

	acceptTNC()
	selectState(*pstate)
	traverseSchools(jump2Breakpoint())
}

func acceptTNC() {
	checkbox, _ := browser.FindElement(selenium.ByID, "checkBoxTou")
	acceptBtn, _ := browser.FindElement(selenium.ByID, "acceptButton")
	checkbox.Click()
	acceptBtn.Click()
}

func selectState(state string) {
	targetState := "state-" + strings.ToLower(state)
	browser.SetImplicitWaitTimeout(5 * time.Second)
	dropdown, _ := browser.FindElement(selenium.ByID, "dropdown-2")
	dropdown.Click()
	states, _ := dropdown.FindElements(selenium.ByTagName, "li")

	for _, s := range states {
		label, _ := s.FindElement(selenium.ByTagName, "label")
		curr_state, _ := label.GetAttribute("for")
		if strings.ToLower(curr_state) == targetState {
			span, _ := label.FindElement(selenium.ByTagName, "span")
			span.Click()
			break
		}
	}
	goBtn, _ := browser.FindElement(selenium.ByID, "go")
	goBtn.Click()
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

	cols, err := element.FindElements(selenium.ByXPATH, "./span/table/tbody/tr[@class='sim-all-row']/td")
	if err == nil && len(cols) >= 2 {
		allDiv, err := cols[1].FindElement(selenium.ByXPATH, "./span[@class='sim-avg']")
		if err == nil {
			allAttr, err := allDiv.Text()
			if err == nil {
				allAvg, _ = strconv.Atoi(allAttr)
			}
		}

		simDiv, err := cols[0].FindElement(selenium.ByXPATH, "./span[@class='sim-avg']")
		if err == nil {
			simAttr, err := simDiv.Text()
			if err == nil {
				simAvg, _ = strconv.Atoi(simAttr)
			}
		}

		simsDiv, err := cols[0].FindElement(selenium.ByXPATH, "./span[@class='err']")
		if err == nil {
			simsAttr, err := simsDiv.Text()
			if err == nil {
				sims := strings.Split(simsAttr, " - ")
				simLow, _ = strconv.Atoi(sims[0])
				simHigh, _ = strconv.Atoi(sims[1])
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
			browser.ExecuteScript("window.open()", nil)
			hanldes, _ := browser.WindowHandles()
			browser.SwitchWindow(hanldes[len(hanldes)-1])
			browser.Get(MYSCHOOL_URL + link)

			browser.SetImplicitWaitTimeout(300 * time.Second)
			topSection, _ := browser.FindElement(selenium.ByCSSSelector, ".topsection-wrapper")
			headerSection, _ := topSection.FindElement(selenium.ByTagName, "h1")
			headerText, _ := headerSection.Text()
			headers := strings.Split(headerText, ",")

			schoolName := strings.Trim(headers[0], " ")
			suburb := strings.Trim(headers[1], " ")
			state := strings.Trim(headers[2], " ")

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
			// naplanMenuBase, _ := browser.FindElement(selenium.ByCSSSelector, "ul.flex.w-100.dropdown-men")
			// naplanMenuItems, _ := naplanMenuBase.FindElements(selenium.ByTagName, "li")
			// naplanMenuItems[1].Click()
			// naplanSubMenuItems, _ := naplanMenuItems[1].FindElements(selenium.ByTagName, "li")
			// naplanSubMenuItems[0].Click()

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
					})
				}

				db.CreateInBatches(&scores, 10)
			}
			util.SaveBreakpoint(page, link_index)

			browser.Close()
			handles, _ := browser.WindowHandles()
			browser.SwitchWindow(handles[0])
			time.Sleep(time.Second)
			link_index += 1
		}

		pages, _ := browser.FindElement(selenium.ByCSSSelector, ".pagination")
		aTags, _ := pages.FindElements(selenium.ByTagName, "a")
		if len(aTags) == 0 {
			break
		}

		nextPage := aTags[len(aTags)-1]
		arrows, _ := nextPage.FindElements(selenium.ByXPATH, "./i[@class='pag_arrow_right']")
		if len(arrows) > 0 {
			href, _ := nextPage.GetAttribute("href")
			browser.Get(href)
			browser.WaitWithTimeoutAndInterval(func(driver selenium.WebDriver) (bool, error) {
				resultSection, _ := driver.FindElement(selenium.ByCSSSelector, ".showing-results")
				return resultSection.IsDisplayed()
			}, 300*time.Second, 200*time.Millisecond)
		}
		page += 1
		link_index = 0
	}
}