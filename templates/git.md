### Browse on Gitlab: [link]({{.Project.HTTPURLToRepo}})

```shell
git clone {{.Project.HTTPURLToRepo}}
cd {{.Project.Name}}
git switch {{.WorkBranchName}}

# Start working...
git status

git add .

# Commit your changes...
git commit -m "Initial Commit"
# Push your work...
git push origin {{.WorkBranchName}}
```

Respond to this page once you're done.
