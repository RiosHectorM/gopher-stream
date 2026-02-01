# 🐹 GopherStream

Sistema de seguimiento de activos en tiempo real construido en **Go**. Diseñado para manejar alta concurrencia y garantizar que ningún dato se pierda, incluso ante caídas de infraestructura.

## 🚀 Características
- **Worker Pool:** Procesamiento asíncrono de eventos para respuestas ultra rápidas.
- **Resiliencia Blindada:** Implementación de Dead Letter Queue (DLQ) en DB y sistema de emergencia en disco local.
- **Auto-Healing:** Monitor que recupera automáticamente datos de fallos previos cuando la DB vuelve a estar online.
- **Arquitectura Hexagonal:** Código desacoplado y fácil de testear.
- **Dockerizado:** Despliegue completo con un solo comando.

## 🛠️ Tecnologías
- **Go** (1.23+)
- **PostgreSQL** (Persistencia)
- **Docker & Docker Compose** (Orquestación)
- **Slog** (Logs estructurados)

## 🚦 Instalación y Uso

1. **Clonar el repositorio:**
   ```bash
   git clone https://github.com/RiosHectorM/gopher-stream
   cd gopher-stream
2. **Configurar variables de entorno:**
   ```bash
    cp .env.example .env
    # Editá el .env con tus credenciales
   ```


3. **Levantar la infraestructura:**
    ```bash
    docker-compose up --build -d
     ```


4. **Probar la API (Postman/cURL):**
* **URL:** `POST http://localhost:8080/tracking`
* **Header:** `X-API-KEY: tu_llave`
* **Body (JSON):**
    ```json
    {
      "asset_id": "TRUCK-01",
      "lat": -31.4135,
      "long": -64.1811,
      "payload": "Ruta 9 - Km 400"
    }
    ```

## 📁 Estructura del Proyecto

* `/cmd`: Punto de entrada de la aplicación.
* `/internal/domain`: Lógica de negocio y modelos.
* `/internal/adapters`: Implementaciones de DB y transporte (HTTP).
* `/storage`: Almacenamiento de logs y recuperación de emergencia.
