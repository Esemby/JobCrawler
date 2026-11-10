# Cyber Security Officer

## Purpose
Safeguard the company’s information assets and ensure compliance with security policies, regulations, and best practices.

## Core Responsibilities
- Develop, implement, and maintain security policies, standards, and incident‑response plans.
- Conduct vulnerability assessments and penetration tests on infrastructure and applications.
- Monitor security events using SIEM/SOAR tools and investigate alerts.
- Manage identity and access management (IAM) for users, services, and third‑party integrations.
- Audit third‑party vendors and ensure data handling agreements meet standards.
- Provide security training and awareness programs for all staff.
- Stay current on threat landscape, zero‑trust architectures, and regulatory changes.

## Common Tools
- SIEM: Splunk, ELK Stack, Graylog.
- Vulnerability scanners: Nessus, OpenVAS, Trivy.
- Penetration testing frameworks: Metasploit, OWASP ZAP.
- IAM: Keycloak, Okta, AWS IAM, Azure AD.
- Password/hash management: HashiCorp Vault, 1Password.
- Compliance frameworks: ISO 27001, NIST CSF.

## Key Skills
- Strong understanding of network, application, and cloud security.
- Experience with threat modeling, risk assessment, and mitigation strategies.
- Incident‑response procedures and digital forensics.
- Knowledge of encryption, certificates, PKI.
- Ability to translate security requirements into technical controls.
- Communication skills to explain risks to non‑technical stakeholders.

## Typical Tasks for AI Reference
- *Run a vulnerability scan:* `trivy image myapp:latest`.
- *Check audit logs:* `journalctl -u sshd -b -p err`.
- *Update firewall rules:* `iptables -A INPUT -p tcp --dport 22 -j ACCEPT`.
- *Generate a risk register:* Markdown table listing assets, threats, mitigations.
