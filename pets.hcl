pet "Ink" {
  type = "cat"
  characteristics {
    sound = "${env.CAT_SOUND}s ${random("evilly", "lazily", "sleepily", "gracefully")}"
  }
}

pet "Swinney" {
  type = "dog"
  characteristics {
    breed = "Dachshund"
  }
}
