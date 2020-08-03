version 1.0

workflow SayHello {

    input {
        Array[String] names = ["Earth", "Mars", "Saturn"]
    }

    scatter(name in names) {
      call Hello {
        input:
            name = name
      }
    }

    output {
        Array[String] msgs = Hello.msg
    }
}

task Hello {

  input {
    String name
  }

  command {
    echo Hello ~{name}! > out.txt
  }

  output {
    String msg = read_string("out.txt")
  }

  runtime {
    docker: "debian:buster-slim"
  }
}