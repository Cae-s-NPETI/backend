# Backend microservices for SledAway

The monorepo backend microservices for **SledAway** – a simulated ride-sharing application.

DISCLAIMER:
> SledAway is in no way a registered trademark or an official application. It is used for educational purposes to simulate a ride-sharing application with modern practices, e.g. microservices. Author bears no responsibility for any injuries caused, cats killed in the using of any portion the app.

# Setting up

## Manually

Prerequisites:
- [Golang](https://go.dev) `>=1.17`

There are two MySQL databases to host. You may choose to use MySQL Workbench, XAMPP, or any other alternatives – but they may require edits to the microservices source code to connect to the DBs properly. This has been tested with XAMPP. Import the `sql_init/01-init_databases.sql` database from the root directory of the source code. This creates the following databases:
- `etia1account`
- `etia1tripmanagement`

Each microservice is contained in subdirectories. To run an individual service, change directory to that microservice, then use `go run .`. This will install needed dependencies and run the service. e.g.
```bash
cd accountManagement
go run .
```
Repeat the same step for all microservices. If you use the [frontend](https://github.com/Cae-s-NPETI/frontend) application, you can view the health status of all three services from the interface.

## Docker

Deploying the backend with Docker Compose is really simple. Database set up is also automatically done on the first run. Simply spin up the following command:
```bash
docker=compose up -d
```

## Summary of microservices

|      | accountManagement |
| ---- | ---- |
| **Description** | Account management microservice for passengers and drivers. |
| **REST Port** | 21801 |
| **Database name** | etia1account |
| **Docker image repository** | caengnp/etia1_accountmanagement |

|      | tripHistory |
| ---- | ---- |
| **Description** | Trip history microservice for logging and retrieving of passenger trips. |
| **REST Port** | 21802 |
| **Database name** | etia1tripmanagement |
| **Docker image repository** | caengnp/etia1_triphistory |

|      | tripManagement |
| ---- | ---- |
| **Description** | Core trip management microservice. |
| **REST Port** | 21803 |
| **Database name** | etia1tripmanagement |
| **Docker image repository** | caengnp/etia1_tripmanagement |
