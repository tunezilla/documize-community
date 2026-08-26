/* fork of documize edition */

-- MD5 for attachments
ALTER TABLE dmz_doc_attachment
    ADD COLUMN c_md5 CHAR(24);
COMMENT ON COLUMN dmz_doc_attachment.c_md5 is 'base64-encoded md5';
CREATE INDEX idx_doc_attachment_md5 ON dmz_doc_attachment (c_md5);
