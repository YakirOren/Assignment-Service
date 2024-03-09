## Hive OnNewAssignment Service

### see example.http

**POST** http://localhost:3000/

1. This will create a subgroup with the for the user (if it doesn't exist already)
2. Forks the `source_repo` and create a new repo named `new_repo_name`
3. Adds the user to the new repo