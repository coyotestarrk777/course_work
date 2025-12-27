#  Сборка бинарного файла ---
FROM golang:1.21-bookworm AS builder

# Установка зависимостей для компиляции Fyne
RUN apt-get update || true && \
    apt-get install -y --no-install-recommends \
    libgl1-mesa-dev \
    xorg-dev \
    gcc \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Сначала копируем зависимости для кэширования слоев
COPY go.mod go.sum ./
RUN go mod download

# Копируем все файлы проекта
COPY . .

# Собираем приложение 
RUN CGO_ENABLED=1 GOOS=linux go build -o music-app .

# Создание чистого образа для запуска ---
FROM ubuntu:22.04

# Предотвращаем зависание установки на выборе часового пояса
ENV DEBIAN_FRONTEND=noninteractive

# Устанавливаем библиотеки рантайма для графики и звука
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    libgl1-mesa-glx \
    libgl1-mesa-dri \
    libx11-6 \
    libxcursor1 \
    libxinerama1 \
    libxrandr2 \
    libxi6 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Копируем только готовый бинарник из сборщика
COPY --from=builder /app/music-app .

# Указываем, что приложение будет использовать графический дисплей
ENV DISPLAY=:0

CMD ["./music-app"]