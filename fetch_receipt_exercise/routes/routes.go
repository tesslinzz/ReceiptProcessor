package routes

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
)

//////////////////////////////////////////////////////////////////////////////////////////////

// GET when starting on localhost:8000
func startApp(c *gin.Context) {
	fmt.Println("Receipt Exercise Started!")

	// Define receipt filenames for POST commands
	receipts := []string{"receipt1.json", "receipt2.json", "receipt3.json", "receipt4.json"}

	// Generate POST commands dynamically for each receipt
	var postCommands []gin.H
	for _, receipt := range receipts {
		postCommands = append(postCommands, gin.H{
			"description": fmt.Sprintf("POST command for processing %s", receipt),
			"example":     fmt.Sprintf(`curl localhost:8000/receipts/process --include --header 'Content-Type: application/json' -d @receipts/%s --request POST`, receipt), // Using single quotes for POST
		})
	}

	// Define GET command
	getCommand := gin.H{
		"description": "GET command with receipt ID to retrieve points",
		"example":     "curl http://localhost:8000/receipts/{id}/points",
	}

	// Return JSON response with POST and GET commands
	c.JSON(http.StatusOK, gin.H{
		"message": "Receipt Exercise Started!",
		"commands": gin.H{
			"POST": postCommands,
			"GET":  []gin.H{getCommand},
		},
	})
}

//////////////////////////////////////////////////////////////////////////////////////////

// POST for new receipt

// initialize a map, key will be the receipt's id and the value will be the points
var pointsMap = make(map[string]int)

// define the Item structure
type Item struct {
	ShortDescription string `json:"shortDescription" binding:"required"`
	Price            string `json:"price" binding:"required,regexp=^[0-9]+\\.[0-9]{2}$"`
}

// define the Receipt structure
type Receipt struct {
	Retailer     string `json:"retailer" binding:"required"`
	PurchaseDate string `json:"purchaseDate" binding:"required"`
	PurchaseTime string `json:"purchaseTime" binding:"required"`
	Total        string `json:"total" binding:"required"`
	Items        []Item `json:"items" binding:"required,min=1"`
}

func validateReceipt(receipt Receipt) error {
	// check the regular expression pattern for the content of receipt, base onn the yaml file

	// check retailer
	retailerPattern := regexp.MustCompile(`^[\w\s\-&]+$`)
	if !retailerPattern.MatchString(receipt.Retailer) {
		return fmt.Errorf("invalid retailer")
	}

	// check total
	decimalRegex := regexp.MustCompile(`^[0-9]+\.[0-9]{2}$`)
	if !decimalRegex.MatchString(receipt.Total) {
		return fmt.Errorf("invalid total")
	}

	// check date
	if _, err := time.Parse("2006-01-02", receipt.PurchaseDate); err != nil {
		return fmt.Errorf("invalid date")
	}

	// check time
	if _, err := time.Parse("15:04", receipt.PurchaseTime); err != nil {
		return fmt.Errorf("invalid time")
	}

	// check item(shortDescription and price)
	itemDescPattern := regexp.MustCompile(`^[\w\s\-]+$`)
	for _, i := range receipt.Items {
		if !itemDescPattern.MatchString(i.ShortDescription) {
			return fmt.Errorf("invalid description")
		}
		if !decimalRegex.MatchString(i.Price) {
			return fmt.Errorf("invalid price")
		}
	}

	return nil
}

func generateReceiptID(receipt Receipt) (string, error) {

	jsonData, err := json.Marshal(receipt)
	if err != nil {
		return "", fmt.Errorf("failed to serialize receipt: %w", err)
	}

	// Compute the SHA-256 hash
	hash := sha256.Sum256(jsonData)

	// Return the hash as a hexadecimal string
	return fmt.Sprintf("%x", hash), nil
}

func calculatePoints(id string, receipt Receipt) {
	// don't need to recaculate if already in map
	if _, exists := pointsMap[id]; exists {
		return
	}

	pts := 0

	// remove characters that aren't letters or digits, and then calculates the length
	filteredRetailer := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) || unicode.IsLetter(r) {
			return r
		}
		return -1 // remove non-alphanumeric characters
	}, receipt.Retailer)

	pts += len(filteredRetailer)

	// if total is round: +50
	totalFloat, _ := strconv.ParseFloat(receipt.Total, 64)
	if totalFloat == math.Trunc(totalFloat) {
		pts += 50
	}

	// if total is multiple of 0.25: +25
	if math.Mod(totalFloat, 0.25) == 0 {
		pts += 25
	}

	// 5 points for each pair
	pts += 5 * (len(receipt.Items) / 2)

	// If the trimmed length of the item description is a multiple of 3: + multiply the price by 0.2(round up to the nearest integer)
	for _, item := range receipt.Items {
		trimmedDesc := strings.TrimSpace(item.ShortDescription)
		length := len(trimmedDesc)
		price, _ := strconv.ParseFloat(item.Price, 64)
		if length%3 == 0 {
			pts += int(math.Ceil(price * 0.2))
		}
	}

	// if odd day: + 6
	dayStr := receipt.PurchaseDate[len(receipt.PurchaseDate)-2:]
	day, err := strconv.Atoi(dayStr)
	if err != nil {
		return
	}
	if day%2 == 1 {
		pts += 6
	}

	// if after 2:00pm and before 4:00pm: + 10
	timeParts := strings.Split(receipt.PurchaseTime, ":")
	hour, _ := strconv.Atoi(timeParts[0])
	min, _ := strconv.Atoi(timeParts[1])
	if (hour == 14 && min > 0) || (hour > 14 && hour < 16) {
		pts += 10
	}

	// put the id, total points as key-value pair into the map
	pointsMap[id] = pts

}

func processReceipt(c *gin.Context) {

	var receipt Receipt

	// Validate receipt: check whether the input json follow the Receipt structure, if could not-> invalid receipt
	if err := c.ShouldBindJSON(&receipt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"BadRequest": "The receipt is invalid."})
		return
	}

	hash, err := generateReceiptID(receipt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"BadRequest": "The receipt is invalid."})
		return
	}

	errFormat := validateReceipt(receipt)

	if errFormat != nil {
		c.JSON(http.StatusBadRequest, gin.H{"BadRequest": "The receipt is invalid."})
		return
	}

	calculatePoints(hash, receipt)

	c.JSON(http.StatusOK, gin.H{"id": hash})

}

// ///////////////////////////////////////////////////////////////////

// GET receipt's points base on the computed hashID

func getPoints(c *gin.Context) {

	// get the input id
	id := c.Param("id")

	// check whether it's a seen receipt
	pts, exist := pointsMap[id]
	if !exist {
		c.JSON(http.StatusNotFound, gin.H{"NotFound": "No receipt found for that ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"points": pts})
}

func RegisterRoutes(router *gin.Engine) {

	// GET route when starting the application
	router.GET("/", startApp)

	// POST route to process the receipt
	router.POST("/receipts/process", processReceipt)

	// GET route to get the points for each receipt(base on ID)
	router.GET("/receipts/:id/points", getPoints)
}
