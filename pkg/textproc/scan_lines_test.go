package textproc

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanLines(t *testing.T) {
	t.Parallel()

	text := `To be successful in this role you will-
    Have relevant experience and a Bachelor's diploma in Computer Science or its equivalent
    Have relevant experience as system or platform engineer with focus on cloud services
    Have experience with SQL and software development using at least 2 out of Golang, Python, Java, C/C++, JavaScript is nice to have
    Have experience with distributed systems and Linux networking, including TCP/IP, SSH, SSL and HTTP protocols
    Possess experience with contemporary DevOps practices and CI/CD tools like Helm, Ansible, Terraform, Puppet, and Chef
    Possess experience with Observability, Performance Analytics and Security tools like Prometheus, CloudWatch, ELK, Sumologic and DataDog
    Have experience with massive data platforms (Hadoop, Spark, Kafka, etc) and design principles (Data Modeling, Streaming vs Batch processing, Distributed Messaging, etc)`

	lines := doScan(text)

	total := 0
	for range lines {
		total++
	}
	require.Equal(t, 8, total)
}

func TestScanLinesManyTexts(t *testing.T) {
	texts := []string{
		`
    2-3+ years of hands-on/production experience.

    Experience with cloud providers (AWS preference).

    Strong experience with Kubernetes and Helm (EKS preference).

    Strong experience with Infrastructure as Code (e.g., Terraform, Terragrunt).

    Experience with logging and monitoring solutions (e.g., Prometheus, Grafana, Elastic Stack).

    GMT+3 time zone +/- 2 hours.

    Proficiency in the software development lifecycle according to DevOps concept, with the ability to write clear, maintainable code in Groovy, Python, Golang, or JavaScript.

    Strong sense of ownership and accountability for your work, with a focus on building trusted relationships with colleagues.

    Strategic thinking and long-term planning to deliver sustainable and scalable solutions.

    Autonomy in learning and executing short-term tasks without compromising quality, requiring minimal mentoring involvement and quickly adapting to new challenges.
	`,
		`
This is a unique Software Engineer role on our Site Reliability Engineering/Security team.

You'll ensure our innovative SaaS/IaaS cloud testing product is up and running for our customers. You'll support observability for our internal and external systems.

When off the shelf products/open source solutions don't exist for a task, you'll build it!

You'll automate things to make your life easier in this role.

The team primarily uses Python, Golang, Kubernetes, Terraform, and GitlabCI. We utilize OpenTelemetry, Prometheus, Grafana, Loki, and Tempo among other observability related tools as well.

Job Type: Full-time

Pay: 185,000.00zł - 285,000.00zł per year

Work Location: Hybrid remote in 00-850 Warszawa

Expected Start Date: 06/01/2025
`,
		`
Your main tasks will be:

    Designing and developing scalable backend solutions using Go and Python.
    Building and maintaining high-load real-time systems with a focus on performance and reliability.
    Writing clean, testable code and implementing comprehensive tests.
    Collaborating with cross-functional teams to influence project strategy and direction.
    Debugging and resolving issues across all software levels and ensuring system stability.
    Managing and optimizing relational (PostgreSQL) and non-relational (Redis) databases.
    Integrating with message brokers such as Kafka and maintaining observability using tools like Prometheus and Grafana.
    Leveraging Kubernetes, Docker, AWS, and other cloud technologies to deploy and scale applications.
    Applying SOLID principles and design patterns to develop robust solutions.

We expect from you:

Technical Expertise:

    At least 3 years of commercial development experience.
    Proficiency in Go (3+ years) and Python
    Solid understanding of design patterns and SOLID principles.
    Experience with relational (PostgreSQL) and non-relational (Redis) databases.
    Familiarity with message brokers (Kafka) and observability tools (e.g., Prometheus, Grafana).
    Hands-on experience with Kubernetes, Docker, and AWS
    Knowledge of CI/CD principles and experience building tested, maintainable code
    Ability to understand and improve others' code while ensuring your own is clear and maintainable
    Experience with ML applications is a plus.
    Soft Skills:
    At least 3 years of commercial development experience.
    Proficiency in Go (3+ years) and Python.
    Solid understanding of design patterns and SOLID principles.
    Experience with relational (PostgreSQL) and non-relational (Redis) databases.
    Familiarity with message brokers (Kafka) and observability tools (e.g., Prometheus, Grafana).
    Hands-on experience with Kubernetes, Docker, and AWS.
    Knowledge of CI/CD principles and experience building tested, maintainable code.
    Ability to understand and improve others' code while ensuring your own is clear and maintainable.
    Experience with ML applications is a plus.
    Strong collaboration skills to work effectively with team members and stakeholders.
    Ability to mentor and learn from others through discussions and code reviews.
    Problem-solving mindset with attention to detail and system-wide understanding.
`,
	}

	lines := ScanLines(texts...)

	for line := range lines {
		fmt.Println(line)
	}
}
