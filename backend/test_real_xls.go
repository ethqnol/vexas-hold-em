package main

import (
	"fmt"
	"os"

	"github.com/extrame/xls"
)

func main() {
	f, err := os.Open("/Users/ewu/Downloads/RE-V5RC-26-3601-Teams-2026-03-10.xls")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()

	xlFile, err := xls.OpenReader(f, "utf-8")
	if err != nil {
		fmt.Println("Error parsing file:", err)
		return
	}

	for s := 0; s < xlFile.NumSheets(); s++ {
		sheet := xlFile.GetSheet(s)
		if sheet == nil {
			continue
		}
		fmt.Printf("Sheet %d MaxRow: %d\n", s, sheet.MaxRow)
		for i := 0; i < 5 && i <= int(sheet.MaxRow); i++ {
			row := sheet.Row(i)
			if row == nil {
				fmt.Printf("Row %d is nil\n", i)
				continue
			}
			fmt.Printf("Row %d LastCol: %d\n", i, row.LastCol())
			fmt.Printf("Row %d Col 0: %q, Col 1: %q, Col 2: %q, Col 3: %q, Col 4: %q\n",
				i, row.Col(0), row.Col(1), row.Col(2), row.Col(3), row.Col(4))
		}
	}
}
