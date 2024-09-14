## [v0.0.1] - 2024-09-13
 
## Load Test
- **output file:** rinhabackendcrebitossimulation-20240913101703340
- **commit:** 774f5640faf4b0453045d940dcaf28ab20ee836f
  
### Added
- Middleware to timeout the requests.
- Monitor for `pgxpool` because of the hanging connections.

### Changed
- Use only one method of repository to execute all transactions. Saving the `pgx.Tx` in the repository was causing inconsistency.

### Fixed
- Database consistency given it was causing `deadlock` or `hanging connection in the pool`.


## [v0.0.2] - 2024-09-14
 
## Load Test
- **output file:** rinhabackendcrebitossimulation-20240914095732829
- **commit:** 5b5ce75c3ac62d60c8fbc306d0d4db825a470e3d

Gatling:
```bash
================================================================================
---- Global Information --------------------------------------------------------
> request count                                      61503 (OK=61503  KO=0     )
> min response time                                      0 (OK=0      KO=-     )
> max response time                                   1046 (OK=1046   KO=-     )
> mean response time                                     4 (OK=4      KO=-     )
> std deviation                                         29 (OK=29     KO=-     )
> response time 50th percentile                          2 (OK=2      KO=-     )
> response time 75th percentile                          2 (OK=2      KO=-     )
> response time 95th percentile                          5 (OK=5      KO=-     )
> response time 99th percentile                         15 (OK=15     KO=-     )
> mean requests/sec                                250.012 (OK=250.012 KO=-     )
---- Response Time Distribution ------------------------------------------------
> t < 800 ms                                         61494 (100%)
> 800 ms <= t < 1200 ms                                  9 (  0%)
> t >= 1200 ms                                           0 (  0%)
> failed                                                 0 (  0%)
================================================================================
```

![Gatling Image Results](./v.0.0.2-gatling-result.png)

### Added
- Script to calculate the results from rinha in my local machine.
- Add CPU and memory limitations
- Postgres config for memory limits

### Changed
- Remove `Serialazable` from the begin on the transaction.
- Docker and Makefile for local and production
- Improve tests for later

### Fixed
