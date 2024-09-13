## [v0.0.1] - 2024-09-06
 
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