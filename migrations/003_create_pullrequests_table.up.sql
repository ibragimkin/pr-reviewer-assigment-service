CREATE TABLE pull_requests (
   pull_request_id     TEXT PRIMARY KEY,
   pull_request_name   TEXT        NOT NULL,
   author_id           TEXT        NOT NULL,
   status              TEXT        NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
   assigned_reviewers  TEXT[]      NOT NULL DEFAULT '{}',
   created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   merged_at           TIMESTAMPTZ NULL,

   CONSTRAINT fk_pull_requests_author
       FOREIGN KEY (author_id)
           REFERENCES users(user_id)
           ON UPDATE CASCADE
           ON DELETE RESTRICT,

-- максимум 2 ревьювера
   CONSTRAINT assigned_reviewers_max_2
       CHECK (cardinality(assigned_reviewers) <= 2)
);
