pet "Spot" {
  type = "dog"
  characteristics {
    breed = "${random(2, "/", "Lab", "Dachshund", "Pug")} Mix"
  }
}
