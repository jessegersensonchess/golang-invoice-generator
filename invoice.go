package main

// go install github.com/creasty/defaults@latest
// go install github.com/angelodlfrtr/go-invoice-generator@latest
// go get github.com/creasty/defaults
// go get github.com/angelodlfrtr/go-invoice-generator

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/spf13/cobra"
)

// Contractor constants
const defaultFromName = "Your Name"
const defaultFromAddress = "Your Address"
const defaultFromContact = "Your Phone"

// Payment constants
const paymentMethod = "Direct deposit"
const routingNumber = "1111111111"
const accountNumber = "222222222"
const accountType = "checking"

// Company constants
const defaultToName = "Company name"
const defaultToAddress = "Company address"
const defaultToContact = "Company phone"

// Input flags
var rate string
var defaultRate = "10.00"
var invoiceNo string
var invoiceDate string
var companyNo string
var fromName string
var fromAddress string
var fromContact string
var toName string
var toAddress string
var toContact string
var taxPercent int

func currentMonthString() string {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, now.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	start := firstOfMonth.Format("Jan 02")
	stop := lastOfMonth.Format("Jan 02, 2006")
	result := fmt.Sprintf("%v - %v", start, stop)
	return result
}
func hoursInCurrentMonth() int {
	// semi random number of hours per month. [1] == January; [12] == December
	myMap := make(map[int]int)
	myMap[1] = 173
	myMap[2] = 168
	myMap[3] = 172
	myMap[4] = 177
	myMap[5] = 174
	myMap[6] = 181
	myMap[7] = 168
	myMap[8] = 179
	myMap[9] = 158
	myMap[10] = 188
	myMap[11] = 169
	myMap[12] = 173
	now := time.Now()
	_, currentMonth, _ := now.Date()
	return myMap[int(currentMonth)]
}

func weekdaysInCurrentMonth() (weekdays int) {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	//currentMonth = month
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, now.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	start := firstOfMonth.Format("Jan 02")
	stop := lastOfMonth.Format("Jan 02, 2006")
	fmt.Println(start, "-", stop)
	weekday := int(firstOfMonth.Weekday())
	for i := firstOfMonth.Day(); i <= lastOfMonth.Day(); i++ {
		dayOfWeek := weekday % 7
		// 0 == sunday; 6 == saturday
		if dayOfWeek != 0 && dayOfWeek != 6 {
			weekdays += 1
		}
		weekday += 1
	}
	return
}

func generateCsv(rate string) {
	// create the file
	f, err := os.Create("/tmp/test.csv")
	if err != nil {
		fmt.Println(err)
	}
	// close the file with defer
	defer f.Close()

	// do operations

	//write directly into file
	f.Write([]byte("Unit,Item,Price"))

	// write a string
	hours := hoursInCurrentMonth()
	f.WriteString("\n" + strconv.Itoa(hours) + ",\"Hours (" + currentMonthString() + ")\"," + rate)
}

var invoiceCmd = &cobra.Command{
	Use:   "generate [CSV file]",
	Short: "Generate invoice from CSV file containing the items for the invoice",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		csvFilePath, err := filepath.Abs(args[0])
		must(err)

		// Check if csv file exists
		if _, err := os.Stat(csvFilePath); os.IsNotExist(err) {
			must(fmt.Errorf("csv file %v does not exists", csvFilePath))
		}

		// Check if the file is indeed csv file
		if filepath.Ext(csvFilePath) != ".csv" {
			must(fmt.Errorf("invalid file type: %v, only accept .csv file", filepath.Ext(csvFilePath)))
		}

		// Read data from csv
		data, err := readDataFromCSV(csvFilePath)
		must(err)

		must(generateInvoice(data))
	},
}

func init() {
	now := fmt.Sprintf("AA-%v", time.Now().Unix())
	invoiceCmd.Flags().StringVarP(&invoiceNo, "invoiceNo", "n", now, "Invoice No., default: <empty>")
	flag.StringVar(&rate, "k", "7", "Hourly pay rate")
	flag.Parse() // after declaring flags we need to call it
	invoiceCmd.Flags().StringVarP(&rate, "defaultRate", "k", rate, "rate, default: 2.03")
	invoiceCmd.Flags().StringVarP(&invoiceDate, "invoiceDate", "d", time.Now().Local().Format("2006-01-02"), "Invoice date in the format of YYYY-MM-DD, default: today's date")
	// companyNo == stick contract ID or leave blank. default 1234
	invoiceCmd.Flags().StringVarP(&companyNo, "companyNo", "p", "1234", "Contract, default: <empty>")
	invoiceCmd.Flags().StringVarP(&fromName, "fromName", "f", defaultFromName, fmt.Sprintf("From name, default: %v", defaultFromName))
	invoiceCmd.Flags().StringVarP(&fromAddress, "fromAddress", "a", defaultFromAddress, fmt.Sprintf("From address, default: %v", defaultFromAddress))
	invoiceCmd.Flags().StringVarP(&fromContact, "fromContact", "c", defaultFromContact, fmt.Sprintf("From contact, default: %v", defaultFromContact))
	invoiceCmd.Flags().StringVarP(&toName, "toName", "o", defaultToName, fmt.Sprintf("To name, default: %v", defaultToName))
	invoiceCmd.Flags().StringVarP(&toAddress, "toAddress", "r", defaultToAddress, fmt.Sprintf("To address, default: %v", defaultToAddress))
	invoiceCmd.Flags().StringVarP(&toContact, "toContact", "t", defaultToContact, fmt.Sprintf("To contact, default: %v", defaultToContact))
	invoiceCmd.Flags().IntVarP(&taxPercent, "taxPercent", "e", 0, "Tax percentage, default: 5%")

	generateCsv(rate)
	rootCmd.AddCommand(invoiceCmd)
}

func generateInvoice(data [][]string) error {
	//	rand.Seed(time.Now().Unix())
	year := time.Now().Year()
	month := int(time.Now().Month())
	//randomNumber := time.Now().Unix()
	marginX := 10.0
	marginY := 20.0
	gapY := 2.0
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(marginX, marginY, marginX)
	pdf.AddPage()
	pageW, _ := pdf.GetPageSize()
	safeAreaW := pageW - 2*marginX

	pdf.ImageOptions("assets/logo.png", 0, 0, 65, 25, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
	pdf.SetFont("Arial", "B", 16)
	_, lineHeight := pdf.GetFontSize()
	currentY := pdf.GetY() + lineHeight + gapY
	pdf.SetXY(marginX, currentY)
	pdf.Cell(40, 10, fromName)

	if companyNo != "" {
		pdf.SetFont("Arial", "BI", 12)
		_, lineHeight = pdf.GetFontSize()
		pdf.SetXY(marginX, pdf.GetY()+lineHeight+gapY)
		pdf.Cell(40, 10, fmt.Sprintf("Contract : %v", companyNo))
	}

	leftY := pdf.GetY() + lineHeight + gapY
	// Build invoice word on right
	pdf.SetFont("Arial", "B", 32)
	_, lineHeight = pdf.GetFontSize()
	pdf.SetXY(130, currentY-lineHeight)
	pdf.Cell(100, 40, "INVOICE")

	newY := leftY
	if (pdf.GetY() + gapY) > newY {
		newY = pdf.GetY() + gapY
	}

	newY += 10.0 // Add margin

	pdf.SetXY(marginX, newY)
	pdf.SetFont("Arial", "", 12)
	_, lineHeight = pdf.GetFontSize()
	lineBreak := lineHeight + float64(1)

	// Left hand info
	splittedFromAddress := breakAddress(fromAddress)
	for _, add := range splittedFromAddress {
		pdf.Cell(safeAreaW/2, lineHeight, add)
		pdf.Ln(lineBreak)
	}
	pdf.SetFontStyle("I")
	pdf.Cell(safeAreaW/2, lineHeight, fmt.Sprintf("Tel: %s", fromContact))
	pdf.Ln(lineBreak)
	pdf.Ln(lineBreak)
	pdf.Ln(lineBreak)

	pdf.SetFontStyle("B")
	pdf.Cell(safeAreaW/2, lineHeight, "Bill To:")
	pdf.Line(marginX, pdf.GetY()+lineHeight, marginX+safeAreaW/2, pdf.GetY()+lineHeight)
	pdf.Ln(lineBreak)
	pdf.Cell(safeAreaW/2, lineHeight, toName)
	pdf.SetFontStyle("")
	pdf.Ln(lineBreak)
	splittedToAddress := breakAddress(toAddress)
	for _, add := range splittedToAddress {
		pdf.Cell(safeAreaW/2, lineHeight, add)
		pdf.Ln(lineBreak)
	}
	pdf.SetFontStyle("I")
	pdf.Cell(safeAreaW/2, lineHeight, fmt.Sprintf("Tel: %s", toContact))

	endOfInvoiceDetailY := pdf.GetY() + lineHeight
	pdf.SetFontStyle("")

	// Right hand side info, invoice no & invoice date
	invoiceDetailW := float64(30)
	pdf.SetXY(safeAreaW/2+30, newY)
	pdf.Cell(invoiceDetailW, lineHeight, "Invoice No.:")
	pdf.Cell(invoiceDetailW, lineHeight, invoiceNo)
	pdf.Ln(lineBreak)
	pdf.SetX(safeAreaW/2 + 30)
	pdf.Cell(invoiceDetailW, lineHeight, "Invoice Date:")
	pdf.Cell(invoiceDetailW, lineHeight, invoiceDate)
	pdf.Ln(lineBreak)

	// Draw the table
	pdf.SetXY(marginX, endOfInvoiceDetailY+10.0)
	lineHt := 10.0
	const colNumber = 5
	header := [colNumber]string{"No", "Description", "Quantity", "Unit Price ($)", "Price ($)"}
	colWidth := [colNumber]float64{10.0, 75.0, 25.0, 40.0, 40.0}

	// Headers
	pdf.SetFontStyle("B")
	pdf.SetFillColor(200, 200, 200)
	for colJ := 0; colJ < colNumber; colJ++ {
		pdf.CellFormat(colWidth[colJ], lineHt, header[colJ], "1", 0, "CM", true, 0, "")
	}

	pdf.Ln(-1)
	pdf.SetFillColor(255, 255, 255)

	// Table data
	pdf.SetFontStyle("")
	subtotal := 0.0

	for rowJ := 0; rowJ < len(data); rowJ++ {
		val := data[rowJ]
		if len(val) == 3 {
			// Column 1: Unit
			// Column 2: Description
			// Column 3: Price per unit
			unit, _ := strconv.Atoi(val[0])
			desc := val[1]
			pricePerUnit, _ := strconv.ParseFloat(val[2], 64)
			pricePerUnit = math.Round(pricePerUnit*100) / 100
			totalPrice := float64(unit) * pricePerUnit
			subtotal += totalPrice

			pdf.CellFormat(colWidth[0], lineHt, fmt.Sprintf("%d", rowJ+1), "1", 0, "CM", true, 0, "")
			pdf.CellFormat(colWidth[1], lineHt, desc, "1", 0, "LM", true, 0, "")
			pdf.CellFormat(colWidth[2], lineHt, fmt.Sprintf("%d", unit), "1", 0, "CM", true, 0, "")
			pdf.CellFormat(colWidth[3], lineHt, fmt.Sprintf("%.2f", pricePerUnit), "1", 0, "CM", true, 0, "")
			pdf.CellFormat(colWidth[4], lineHt, fmt.Sprintf("%.2f", totalPrice), "1", 0, "CM", true, 0, "")
			pdf.Ln(-1)
		}
	}

	// Calculate the subtotal
	pdf.SetFontStyle("B")
	leftIndent := 0.0
	for i := 0; i < 3; i++ {
		leftIndent += colWidth[i]
	}
	pdf.SetX(marginX + leftIndent)
	pdf.CellFormat(colWidth[3], lineHt, "Subtotal", "1", 0, "CM", true, 0, "")
	pdf.CellFormat(colWidth[4], lineHt, fmt.Sprintf("%.2f", subtotal), "1", 0, "CM", true, 0, "")
	pdf.Ln(-1)

	//	taxAmount := math.Round(subtotal*float64(taxPercent)) / 100
	taxAmount := 0.00
	pdf.SetX(marginX + leftIndent)
	pdf.CellFormat(colWidth[3], lineHt, "Tax", "1", 0, "CM", true, 0, "")
	pdf.CellFormat(colWidth[4], lineHt, fmt.Sprintf("%.2f", taxAmount), "1", 0, "CM", true, 0, "")
	pdf.Ln(-1)

	grandTotal := subtotal + taxAmount
	pdf.SetX(marginX + leftIndent)
	pdf.CellFormat(colWidth[3], lineHt, "Total", "1", 0, "CM", true, 0, "")
	pdf.CellFormat(colWidth[4], lineHt, fmt.Sprintf("%.2f", grandTotal), "1", 0, "CM", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFontStyle("")
	pdf.Ln(lineBreak)
	pdf.Cell(safeAreaW, lineHeight, fmt.Sprintf("Payment Method: %v", paymentMethod))
	pdf.Ln(lineBreak)
	pdf.Cell(safeAreaW, lineHeight, fmt.Sprintf("  -- beneficiary name: %v", defaultFromName))
	pdf.Ln(lineBreak)
	pdf.Cell(safeAreaW, lineHeight, fmt.Sprintf("  -- routing number: %v", routingNumber))
	pdf.Ln(lineBreak)
	pdf.Cell(safeAreaW, lineHeight, fmt.Sprintf("  -- account number: %v", accountNumber))
	pdf.Ln(lineBreak)
	pdf.Cell(safeAreaW, lineHeight, fmt.Sprintf("  -- account type: %v", accountType))

	shortName := strings.Replace(defaultFromName, " ", "", -1)
	filename := fmt.Sprintf("/tmp/%v-%v-%02d.pdf", shortName, year, month)
	return pdf.OutputFileAndClose(filename)
}

func breakAddress(input string) []string {
	var address []string
	const limit = 10
	splitted := strings.Split(input, ",")
	prevAddress := ""
	for _, add := range splitted {
		if len(add) < 10 {
			prevAddress = add
			continue
		}
		currentAdd := strings.TrimSpace(add)
		if prevAddress != "" {
			currentAdd = prevAddress + ", " + currentAdd
		}
		address = append(address, currentAdd)
		prevAddress = ""
	}

	return address
}
