# Etapa 1: Construcción
FROM golang:1.22.5 AS builder

# Establecer el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copiar los archivos del proyecto al contenedor

COPY . .

# Construir el binario
RUN go build -o main .

# Etapa 2: Imagen final
FROM debian:bullseye-slim

# Establecer el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copiar el binario desde la etapa de construcción
COPY --from=builder /app/main .

# Exponer el puerto en el que corre la aplicación (cambiar si es necesario)
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["./main"]
