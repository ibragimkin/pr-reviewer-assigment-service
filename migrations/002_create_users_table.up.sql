CREATE TABLE users (
   user_id   TEXT PRIMARY KEY,
   username  TEXT NOT NULL,
   team_name TEXT NOT NULL,
   is_active BOOLEAN NOT NULL DEFAULT TRUE,

   CONSTRAINT fk_users_team
       FOREIGN KEY (team_name)
           REFERENCES teams(team_name)
           ON UPDATE CASCADE
           ON DELETE RESTRICT
);
