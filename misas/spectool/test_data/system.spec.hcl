system "unit test" {
  description = "System made for unit tests of go MISAS"
  sources = ["."]
}

command "user.register" {
  description = "allows queuing a work item"

  field "id" {
    description = "Optional ID of the work item to be registered. If none is provided, one will be generated."
    type = "identifier"
  }

  meta "gen:go:name" {
    value = "RegisterUserCommand"
  }
}

event "user.registered" {
  description = "allows queuing a work item"

  field "id" {
    description = "ID of the work item that was registered."
    type = "identifier"
  }

  field "registeredAt" {
    description = "date and time at which the work item was registered."
    type = "dateTime"
  }

}