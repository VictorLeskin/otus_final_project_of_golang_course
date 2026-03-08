set HW=chore/project-setup

git switch main
pause

git checkout -b %HW%
# create join the remove and local branchs
git branch --set-upstream-to=origin/%HW% %HW%