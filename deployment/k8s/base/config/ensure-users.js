const targetDbStr = 'auth-service'
db.createUser({
    user: "admin",
    pwd: "changeit",
    roles: [
        {
            role: "userAdminAnyDatabase",
            db: "admin"
        },
        {
            role: "dbAdminAnyDatabase",
            db: "admin"
        },
        {
            role: "readWriteAnyDatabase",
            db: "admin"
        }
    ]
})