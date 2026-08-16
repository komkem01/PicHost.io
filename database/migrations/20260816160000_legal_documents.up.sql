CREATE TABLE IF NOT EXISTS legal_documents (
    key VARCHAR(64) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO legal_documents (key, title, content) VALUES
('terms', 'Terms of Service / ข้อตกลงและเงื่อนไขการให้บริการ', 'By accessing or using PicHost.io, you agree to be bound by these Terms of Service.
การเข้าใช้บริการ PicHost.io ถือว่าท่านได้ยอมรับข้อตกลงและเงื่อนไขการให้บริการนี้แล้ว

1. Acceptance of Terms / การยอมรับข้อตกลง
By accessing or using PicHost.io, you agree to be bound by these Terms of Service.
การเข้าใช้บริการ PicHost.io ถือว่าท่านได้ยอมรับข้อตกลงและเงื่อนไขการให้บริการนี้แล้ว

2. Acceptable Use & Content Policy / นโยบายการใช้งานและเนื้อหา
Illegal content, malware, phishing, and copyrighted content violations are strictly prohibited.
ห้ามใช้อัปโหลดไฟล์ที่ผิดกฎหมาย สื่อลามกอนาจาร ไวรัส/มัลแวร์ หรือไฟล์ละเมิดลิขสิทธิ์

3. Account Termination / การยกเลิกบัญชี
We reserve the right to suspend or terminate accounts that violate our terms of service without prior notice.')
ON CONFLICT (key) DO NOTHING;

INSERT INTO legal_documents (key, title, content) VALUES
('privacy', 'Privacy Policy / นโยบายความเป็นส่วนตัว', 'We collect account email, username, uploaded image files, and security access logs strictly for operating PicHost.io.
เราจัดเก็บอีเมล ชื่อผู้ใช้ ภาพที่ถูกอัปโหลด และประวัติบันทึกความปลอดภัยเฉพาะที่จำเป็นสำหรับการให้บริการเท่านั้น

1. Information We Collect / ข้อมูลที่เราจัดเก็บ
We collect account email, username, uploaded image files, and security access logs strictly for operating PicHost.io.
เราจัดเก็บอีเมล ชื่อผู้ใช้ ภาพที่ถูกอัปโหลด และประวัติบันทึกความปลอดภัยเฉพาะที่จำเป็นสำหรับการให้บริการเท่านั้น

2. Data Protection & PDPA / การคุ้มครองข้อมูลส่วนบุคคล
We protect your data according to Thailand PDPA standards. Your personal data is never sold to third parties.')
ON CONFLICT (key) DO NOTHING;
