set BR=feature/rate-limiter-service
set BR=feature/white-black-lists-service

git switch main
pause

git checkout -b %BR%
# create join the remove and local branchs
git branch --set-upstream-to=origin/%BR% %BR%

git push -u origin feature/%BR%