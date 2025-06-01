# Go Server Starter - 🚨 Work in progress! 🚨

Basic Server Template in Golang

## Description

Go Server Starter is a barebones, simple and easy to understand template, containing sensiable "boiler plate" for a backend.
Project Goals:
* Keep it simple stupid

*Disclamer:* Some of the feature have been added because I wanted to implyment them, they might not be the best choise for production golang backend. I have flagged them so that can be removed or disabled easily.


Current "Features":
* Middlewares "chain builder"
  - Access Logging
  - Panic Recovery

* Database connection
  -  Postgres connection pool



* Repository Pattern
* Postgres database connection
  * Test Container
* Logging - slog and zerolog (currently set up for zerolog - can be changed if the dependancy is a concern)
* Prometheus metrics exporting, Pyroscope Profiling

### Todo:
* Caching
* Documentation
* Json Encode and Decode
* SQLc
* Dockerfile & Docker Compose
* SQLite Support
* Websockets and SSE Event Broker
* Postgres Listener

## 💡Usage
Template/Clone/Fork the repository, customise and enjoy

### Folder
- Database: All database connection related files
- Models: "Things" - also known as entities
- Repository: Database related logic - can take in a transaction
- Services: Appliation logic e.g. Caching, starting a database transaction
- Middleware: Server middleware e.g. Auth - Included access logging and recovery
- logging: Zerologger and slog logger implymentations
- utilities: Those handy bits of code that I never know were to put

## Authors

doug-benn - [github](www.github.com/doug-benn)

## License

This project is licensed under GNU General Public License v3.0 - see the LICENSE file for details

## Acknowledgments

Coming soon