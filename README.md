## Hive Assignment Service (Gitlab Service)

What is this?
* A service that runs an action when a new assignment is given.



## Description:
* adds a description to the assigment from a template


## Gitlab:
### What happens when the service gets a gitlab request?
1. Creates a subgroup for the user (if it doesn't exist already)
2. Forks the `source_repo` and create a new repo named `new_repo_name`
3. Adds the user to the new repo
4. Removes the fork relation
5. Remove branch protection, if `work_branch_name` and `base_branch_name` are the same

### On creation data
```json
{
  "gitlab": {
    "namespace": "students/python",
    "source_repo": "templates/first-template",
    "new_repo_name": "The Cloner",
    "base_branch_name": "main",
    "work_branch_name": "work_branch_name",
    "detailed_instructions": true
  }
}
```

![](images/full_description.png)


* With `detailed_instructions` set to `false`
![](images/mininal_description.png)


## Changing the access level
* By default, the service adds the users as developers to new repositories
* set the env variable `ACCESS_LEVEL=40`, to add them as maintainers.
* https://docs.gitlab.com/ee/api/access_requests.html


### Configuring instructions
* see templates directory
* **NOTE: `git.md` and `short-git.md` files must exist.** 
