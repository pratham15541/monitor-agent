mvn clean package -DskipTests
docker compose down -v  
docker compose up --build