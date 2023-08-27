# Insert Million Data to MySQL
This project to do insert millions data from csv file to MySQL.

This project use method **Worker Pool** and **Failover mechanism**.

This means when the insert data fails, the project will try again do inserted until all data is inserted to DB.

This project inspired from mas **Noval Agung**, you can see the original blog in here: https://dasarpemrogramangolang.novalagung.com/A-fungsi-closure.html.

# How to run
1. Clone this project.
```

```

2. Installation library
```
go mod tidy
```

3. Download file
Download csv file on the link: http://downloads.majestic.com/majestic_million.csv.
Save this file in root project.
*Note: for this file license you can see: https://blog.majestic.com/development/majestic-million-csv-daily/.

4. Create database in your machine, in this case use MySQL
```
create database test;
```

5. Run App to begin import data to database
```
go run main.go
```