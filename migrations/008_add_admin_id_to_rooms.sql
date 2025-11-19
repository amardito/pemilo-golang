-- +migrate Up
-- Add admin_id column to rooms table
ALTER TABLE rooms ADD COLUMN admin_id VARCHAR(36);

-- Add foreign key constraint
ALTER TABLE rooms ADD CONSTRAINT fk_rooms_admin_id 
    FOREIGN KEY (admin_id) REFERENCES admins(id) ON DELETE CASCADE;

-- Create index on admin_id
CREATE INDEX idx_rooms_admin_id ON rooms(admin_id);

-- +migrate Down
ALTER TABLE rooms DROP CONSTRAINT IF EXISTS fk_rooms_admin_id;
ALTER TABLE rooms DROP COLUMN IF EXISTS admin_id;
