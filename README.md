# Go Server Starter - 🚨 Work in progress! 🚨

Basic HTTP Server Template in Golang

## Description

The goal for this project is a simple easy to understand starter HTTP server.


Current "Features":
* Postgres database connection
  * Test Container
* Logging - slog and zerolog (currently set up for zerolog - can be changed if the dependancy is a concern)
* Prometheus metrics server

I want to add:
* Docker Compose
  * Including Grafana & Prometheus
* SQLite Support
* Websockets and SSE Event Broker
* Switch back to slogger
* Access Log & Recovery testing
* Postgres Listener

## 💡Usage
Template/Clone/Fork the repository and enjoy

### Folder
- Models: "Things" - also known as entities
- Repository: Database related logic - can take in a transaction
- Services: Appliation logic e.g. Caching, starting a database transaction

## Authors

doug-benn - [github](www.github.com/doug-benn)

## License

This project is licensed under GNU General Public License v3.0 - see the LICENSE file for details

## Acknowledgments

Coming soon