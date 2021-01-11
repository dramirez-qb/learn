package com.learn.springboot;

import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.bind.annotation.RequestMapping;

@RestController
public class HelloController {

	@RequestMapping("/")
	public String index() {
		return "Hello World from Spring Boot!";
	}

	@RequestMapping("/ping")
	public String ping() {
		return "pong";
	}

	@RequestMapping("/healthz")
	public String healthz() {
		return "Alive from Spring Boot!";
	}

}
