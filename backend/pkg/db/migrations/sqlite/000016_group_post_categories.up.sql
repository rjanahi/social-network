CREATE TABLE IF NOT EXISTS group_post_categories (
  group_post_id INTEGER NOT NULL,
  category_id INTEGER NOT NULL,
  PRIMARY KEY (group_post_id, category_id),
  FOREIGN KEY (group_post_id) REFERENCES group_posts(id) ON DELETE CASCADE,
  FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);