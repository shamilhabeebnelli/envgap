# Research notes

Checked 24 Nov 2026 before building.

- Search for dotenv compare tools returned many existing projects, including https://github.com/shahadulhaider/envdiff (Go, GPL-3.0, "comparing, validating, and syncing .env files"), https://github.com/jjdoor/dotdrift, https://github.com/AtyahsLab/env-doctor, https://github.com/freedom07/envdiff and https://pypi.org/project/envcompare/.
- https://github.com/dotenv-linter/dotenv-linter is an established linter for .env files.
- Demand is therefore real but the space is crowded. This project does not claim to be first or better. Its choice: a minimal, MIT, stdlib-only comparison whose output never contains values, with tests that enforce that.
- Not verified: how many people need value-free output specifically.
