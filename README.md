# pet-sounds

pet-sounds is an application that uses a few intermediate HCL2 concepts to decode a configuration file and output some information about various pets.

`pet.go` and `pets.hcl` contain all the interesting code.

## Polymorphism using Partial Decoding

"Partial Decoding" is an extremely powerful concept in HCL decoding. It allows you to create configurations that are self referential, where one decoding step relies on another.

In our case, we'll use them to make the `pet` block generic, with a `type` field that determines what kind of pet it is. The `characteristics` block inside of `pet` is still type safe, with `dog` and `cat` blocks with unique fields that cannot be used in with wrong type of pet.

## Variables

Variables are also useful for making HCL more dynamic.

In our case, we'll use a nested variable that can load certain information from the environment: `env.CAT_SOUND` for example.

## Functions

Functions can be made that are custom to your HCL decoding domain. These allow for even more complicated and flexible HCL config files.

In our case, we'll create a random function that picks a random string from the inputs provided. Then we'll use that function to express our domain specific knowledge about the fickleness of cats.

## Ink the cat

Ink the cat from the configuration file is a real cat and he loves occupying desk space while you're trying to program.

![Image of Ink](/ink.jpg)
