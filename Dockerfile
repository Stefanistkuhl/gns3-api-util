FROM python:3-alpine

WORKDIR /app

COPY src /app/src
COPY pyproject.toml .

RUN pip install . --break-system-packages

ENTRYPOINT ["/usr/local/bin/gns3util"]
