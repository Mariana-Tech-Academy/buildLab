package main



import (
    "Fhub/models"
    "Fhub/routes"
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/joho/godotenv"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)



func main() {
    // build connection
    godotenv.Load()

    host := os.Getenv("DB_HOST")
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    port := os.Getenv("DB_PORT")
    name := os.Getenv("DB_NAME")
    sslmode := os.Getenv("SSL_MODE")

    connStr := fmt.Sprintf("host=%s user=%s password=%s port=%s dbname=%s sslmode=%s", host, user, password, port, name, sslmode)

    // open database
    db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
    if err != nil {
        log.Fatal("failed to open db: ", err)
    }

    // run migration table
    err = db.AutoMigrate(&)
    if err != nil {
        log.Fatal("failed to automigrate: ", err)
    }

    //  set up router
    r := routes.Router(db)

    //  start the server
    fmt.Println("server is running")
    err = http.ListenAndServe(":8080", r)
    if err != nil {
        log.Fatal("error starting server: ", err)
    }

}