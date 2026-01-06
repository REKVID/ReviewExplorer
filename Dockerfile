FROM golang:1.23

RUN apt-get update && apt-get install -y \
    python3 python3-pip \
    chromium \
    chromium-driver \
    libnss3 \
    libgconf-2-4 \
    libfontconfig1 \
    && rm -rf /var/lib/apt/lists/*
RUN pip3 install selenium mysql-connector-python --break-system-packages

WORKDIR /app
COPY . .
RUN go build -o server backend/main.go

CMD ["./server"]
