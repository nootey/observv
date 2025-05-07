# observv

This repository contains a full implementation of the LGTM (Loki, Tempo, Grafana, Mimir) observability stack.
It uses a very simple HTTP server, built in Go and Gin. 

- I used a minimalistic server, to keep the focus on the observability.
- Having a smaller, less convoluted project, keeps the structure clean and presentable.
- It also allows for faster testing and quicker updates.
- The stack is also meant to be reusable.