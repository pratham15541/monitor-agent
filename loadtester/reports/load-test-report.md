==========================================

LOAD TEST SUMMARY

==========================================

Started: 2026-07-10T22:06:59+05:30
Finished: 2026-07-10T22:20:33+05:30
Duration: 13m33.722586388s

Companies: 3
Systems: 500
Registered: 500
Connected: 8
Authentication Success: 3
Authentication Failure: 0
Peak Concurrent Connections: 500
Heartbeats: 10736
Telemetry Messages: 169338
Telemetry/sec: 208.12
Commands Executed: 18
Reconnect Attempts: 264
Reconnect Success: 760
Average Registration Time: 0s
Average Connection Time: 241.578791ms
Average Latency: 281.9762ms
P95: 1.24232278s
P99: 5.568274012s
Backend CPU: 0.00
Backend Memory: 0.00
Load Tester Memory Samples: 80

Generated Test Data

Company ID: e0582edf-fd67-47ca-8fd1-b2cd8f0f9e3f
Company: Company A
Email: companyA@test.local
Password: Test@123
Systems: 50
Connected: 50
JWT Token: eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJlMDU4MmVkZi1mZDY3LTQ3Y2EtOGZkMS1iMmNkOGYwZjllM2YiLCJpc3MiOiJtb25pdG9yLXRvb2wiLCJpYXQiOjE3ODM3MDE0MjAsImV4cCI6MTc4MzcwNTAyMH0.i5HwjtIaZ6_1cerLdCGDOi3B5qigoT82Q3knOdScK9U
----------------------------------------
Company ID: 87974420-1f99-4e27-8adb-60b82da62cbb
Company: Company B
Email: companyB@test.local
Password: Test@123
Systems: 100
Connected: 100
JWT Token: eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiI4Nzk3NDQyMC0xZjk5LTRlMjctOGFkYi02MGI4MmRhNjJjYmIiLCJpc3MiOiJtb25pdG9yLXRvb2wiLCJpYXQiOjE3ODM3MDE0MjAsImV4cCI6MTc4MzcwNTAyMH0.HXVBcCXEDEs2tSCHOV3KeaE1owKY_aodGiZL8cm1IwM
----------------------------------------
Company ID: 9b2ec6cb-f719-42ab-9318-f55f30b2d667
Company: Company C
Email: companyC@test.local
Password: Test@123
Systems: 350
Connected: 350
JWT Token: eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiI5YjJlYzZjYi1mNzE5LTQyYWItOTMxOC1mNTVmMzBiMmQ2NjciLCJpc3MiOiJtb25pdG9yLXRvb2wiLCJpYXQiOjE3ODM3MDE0MjAsImV4cCI6MTc4MzcwNTAyMH0.gb2xa7mPUxxUBdCaYL_6ParHeYcJ2VyI9xQzGu64qCE
----------------------------------------

Analysis

WARN
- backend metrics unavailable: jvm.memory.used unavailable: GET /actuator/metrics/jvm.memory.used returned 401 
- load tester memory increased materially during the run
- not all agents stayed connected through the test window

Recommendations

- Review goroutine lifetimes and connection shutdown paths
