==========================================

LOAD TEST SUMMARY

==========================================

Started: 2026-07-10T22:04:21+05:30
Finished: 2026-07-10T22:05:50+05:30
Duration: 1m28.356194703s

Companies: 3
Systems: 500
Registered: 500
Connected: 116
Authentication Success: 3
Authentication Failure: 0
Peak Concurrent Connections: 116
Heartbeats: 224
Telemetry Messages: 3157
Telemetry/sec: 35.73
Commands Executed: 0
Reconnect Attempts: 4
Reconnect Success: 120
Average Registration Time: 0s
Average Connection Time: 9.06478ms
Average Latency: 54.845794ms
P95: 132.710568ms
P99: 200.891618ms
Backend CPU: 0.00
Backend Memory: 0.00
Load Tester Memory Samples: 8

Generated Test Data

Company ID: e0582edf-fd67-47ca-8fd1-b2cd8f0f9e3f
Company: Company A
Email: companyA@test.local
Password: Test@123
Systems: 50
Connected: 13
JWT Token: eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJlMDU4MmVkZi1mZDY3LTQ3Y2EtOGZkMS1iMmNkOGYwZjllM2YiLCJpc3MiOiJtb25pdG9yLXRvb2wiLCJpYXQiOjE3ODM3MDEyNjIsImV4cCI6MTc4MzcwNDg2Mn0.FMPUJiStSNSHyxW9O6Ir4kIyXlzE3-riW61lnFNegyQ
----------------------------------------
Company ID: 87974420-1f99-4e27-8adb-60b82da62cbb
Company: Company B
Email: companyB@test.local
Password: Test@123
Systems: 100
Connected: 19
JWT Token: eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiI4Nzk3NDQyMC0xZjk5LTRlMjctOGFkYi02MGI4MmRhNjJjYmIiLCJpc3MiOiJtb25pdG9yLXRvb2wiLCJpYXQiOjE3ODM3MDEyNjIsImV4cCI6MTc4MzcwNDg2Mn0.V5IXQNJKdminZGlOl6wf28t5OSwNIVZ-7Z_M7SJPmA0
----------------------------------------
Company ID: 9b2ec6cb-f719-42ab-9318-f55f30b2d667
Company: Company C
Email: companyC@test.local
Password: Test@123
Systems: 350
Connected: 84
JWT Token: eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiI5YjJlYzZjYi1mNzE5LTQyYWItOTMxOC1mNTVmMzBiMmQ2NjciLCJpc3MiOiJtb25pdG9yLXRvb2wiLCJpYXQiOjE3ODM3MDEyNjIsImV4cCI6MTc4MzcwNDg2Mn0.LLwTnR_j2xqGCNx1OQ-nzgyT26M2-wiH1e0i_N7JJ5g
----------------------------------------

Analysis

WARN
- backend metrics unavailable: jvm.memory.used unavailable: GET /actuator/metrics/jvm.memory.used returned 401 
- load tester memory increased materially during the run
- not all agents stayed connected through the test window

Recommendations

- Review goroutine lifetimes and connection shutdown paths
