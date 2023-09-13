package util

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func LoadBreakpoint() (int, int) {
	bts, err := os.ReadFile("breakpoint")
	if err != nil {
		return 0, 0
	}

	parts := strings.Split(string(bts), ",")
	if len(parts) != 2 {
		return 0, 0
	}

	page, _ := strconv.Atoi(parts[0])
	index, _ := strconv.Atoi(parts[1])

	return page, index
}

func SaveBreakpoint(page, index int) {
	log.Printf("saving breakpoint on page %d, index %d", page, index)
	os.WriteFile("breakpoint", []byte(fmt.Sprintf("%d,%d", page, index)), os.ModePerm)
}
