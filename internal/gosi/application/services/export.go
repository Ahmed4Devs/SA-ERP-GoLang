package services

import (
	"encoding/xml"
	"erp/internal/gosi/domain"
)

type contributionReport struct {
	XMLName     xml.Name         `xml:"urn:erp:gosi:export:v1 GOSIContributionReport"`
	Year        int              `xml:"year,attr"`
	Month       int              `xml:"month,attr"`
	LegalEntity reportEntity     `xml:"LegalEntity"`
	Summary     reportSummary    `xml:"Summary"`
	GeneratedAt string           `xml:"GeneratedAt"`
	Employees   []reportEmployee `xml:"Employees>Employee"`
}
type reportEntity struct {
	ID string `xml:"id,attr"`
}
type reportSummary struct {
	SubscriberCount int    `xml:"subscriberCount,attr"`
	EmployeeTotal   string `xml:"employeeTotal,attr"`
	EmployerTotal   string `xml:"employerTotal,attr"`
	GrandTotal      string `xml:"grandTotal,attr"`
}
type reportEmployee struct {
	ID             string       `xml:"id,attr"`
	Number         string       `xml:"number,attr"`
	EmployeeAmount string       `xml:"employeeAmount,attr"`
	EmployerAmount string       `xml:"employerAmount,attr"`
	Rules          []reportRule `xml:"Rule"`
}
type reportRule struct {
	VersionID      string `xml:"versionId,attr"`
	Code           string `xml:"code,attr"`
	Basis          string `xml:"basis,attr"`
	EmployeeRate   string `xml:"employeeRate,attr"`
	EmployerRate   string `xml:"employerRate,attr"`
	EmployeeAmount string `xml:"employeeAmount,attr"`
	EmployerAmount string `xml:"employerAmount,attr"`
}

func employeeReport(calc domain.ContributionCalculation, lines []domain.CalculationLine) reportEmployee {
	employee := reportEmployee{ID: calc.EmployeeID.String(), Number: calc.EmployeeNumber, EmployeeAmount: calc.TotalEmployeeAmount.String(), EmployerAmount: calc.TotalEmployerAmount.String()}
	for _, line := range lines {
		employee.Rules = append(employee.Rules, reportRule{VersionID: line.RuleVersionID.String(), Code: line.RuleCode, Basis: line.BasisAmount.String(), EmployeeRate: line.EmployeeRate.String(), EmployerRate: line.EmployerRate.String(), EmployeeAmount: line.EmployeeAmount.String(), EmployerAmount: line.EmployerAmount.String()})
	}
	return employee
}
