FROM python:3.13-alpine

WORKDIR /app

COPY src .
COPY pyproject.toml .

RUN pip install . --break-system-packages

ENTRYPOINT ["/usr/local/bin/gns3util"]
