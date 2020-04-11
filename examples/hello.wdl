version 1.0

workflow SayHello {

    input {
        String name = "World"
    }

    call Hello {
        input:
            name = name
    }

    output {
        String msg = Hello.msg
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
    File msg = "out.txt"
  }

  runtime {
    docker: "debian:buster-slim"
  }
}