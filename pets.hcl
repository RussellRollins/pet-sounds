pet "Ink" {
  type = "cat"
  characteristics {
    sound = "${env.CAT_SOUND}s ${random(2, " and ", "evilly", "lazily", "sleepily", "gracefully")}"
  }
}

pet "Swinney" {
  type = "dog"
  characteristics {
    breed = "Dachshund"
  }
}
