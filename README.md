# RANDOM PICKER CVV
Web app used to draw a student name from classmates using CVV networks

## USAGE
To start the webapp

```bash
go run main.go <uid> <password>
```

### Methods
1. `curl -X GET localhost:8000/draw` to random extract one name from the class
2. `curl -X GET localhost:8000/getClass` to get all the names of a class
3. `curl -X POST localhost:8000/addMate -F name=<name>` to add one name to the class
4. `curl -X POST localhost:8000/removeMate -F id=<id>` to remove one name from the class by id (all the ids are shown in /getClass response)


