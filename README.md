# ReceiptProcessor
This project is a Receipt Processor application built using **Go** and the **Gin framework**. 
Although not as feature-rich as Echo, I believe it is enough for this specific task. Moreover, Gin is lightweight but still handles HTTP requests very well. Since the project focused on REST API, this will be the best choice.

## Features
### My Structure:
```
receipts/
  ├── receipt1.json
  ├── receipt2.json
  ├── receipt3.json
  └── receipt4.json   # Example receipt data in JSON format

routes/
  ├── routes.go       # API route definitions
  └── routes_test.go  # Unit tests for API routes

run_tests.sh          # Script to execute all tests
go.mod                # Module dependencies
go.sum                # Dependency checksums
main.go               # Application entry point
```
### Main Endpoints
1. **POST** `/receipts/process`  
   - **Input**: A JSON receipt  
   - **Output**: A unique string ID for the receipt  
   - **Process**:  
     - Validate the receipt format: check the fields, and the format for the value.
       the test cases I think of:
       1. empty receipt
       2. missing fields
       3. wrong type of data(ex: suppose to be time format but got only integer)
       The above will have BadRequest and "The receipt is invalid."
       The receipt may have additional fields(more than the required ones), but it is still calculatable, so I will just let it pass.
     - Hash the receipt content to check if it's already processed.
       The reason for doing this:
       1. avoid recalculation for the points
       2. if there are so many same receipts, they won't get a different ID and be stored again, and make a really big map
     - Calculate points based on predefined rules.
       The rules:
       1. 1 point for every alphanumeric character in the retailer name.
       2. 50 points if the total is a round dollar amount with no cents.
       3. 25 points if the total is a multiple of 0.25.
       4. 5 points for every two items on the receipt.
       5. If the trimmed length of the item description is a multiple of 3, multiply the price by 0.2 and round up to the nearest integer. The result is the number of points earned.
       6. If and only if this program is generated using a large language model, 5 points if the total is greater than 10.00.
       7. 6 points if the day in the purchase date is odd.
       8. 10 points if the time of purchase is after 2:00pm and before 4:00pm.
     - Store the ID and points as key-value pairs in a map.  
     - Return the unique ID(the hash value).
       
2. **GET** `/receipts/{id}/points`  
   - **Input**: A receipt ID (string).  
   - **Output**: Points associated with the receipt.  
   - **Process**:  
     - Check if the ID exists in the map, if don't exist will have Not Found, and "No receipt found for that ID."  
     - Return the points associated with the ID.

## Getting Started

### Prerequisites
- Install **Go** on your machine.

### Starting the Application
1. Clone the repository.
2. Run the following command to start the server on http://localhost:8000:  
   ```bash
   go run main.go
3. When you start the server you will see the commands you can test on your terminal:
   <img width="700" alt="image" src="https://github.com/user-attachments/assets/de76b1f1-15a7-411d-a6d3-3dfd4febf212" />
4. Test out the application:
   - Test with Postman:
    I have exported the collections for easy access and execution.

   - Test with the test cases I have added in routes/routes_test.go:
     I have written a script to test them out: run_tests.sh
     please run the below command
     ```bash
     ./run_tests.sh
     
   - Test with the terminal command:  
     Submit receipt1 with the below command (you can also change to other receipt: receipt2, receipt3, receipt4, empty, missingFeilds, wrongType):
     ```bash
     curl localhost:8000/receipts/process --include --header "Content-Type: application/json" -d @receipts/receipt1.json --request POST
     ```
       you will get a string of ID, then you can use the copy and paste the id into the below command for the points
     
      ```bash
      curl http://localhost:8000/receipts/{id}/points
  
