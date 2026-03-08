set BR=chore/project-setup1

git switch main
pause

git checkout -b %BR%
# create join the remove and local branchs
git branch --set-upstream-to=origin/%BR% %BR%