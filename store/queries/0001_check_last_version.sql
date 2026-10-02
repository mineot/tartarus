SELECT id, versionDate, versionNumber
FROM migrations
ORDER BY versionNumber DESC, id DESC
LIMIT 1;
