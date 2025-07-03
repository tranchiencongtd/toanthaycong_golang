-- Khởi tạo database với dữ liệu mẫu cho nền tảng học tập
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "citext";

-- Thêm các danh mục mẫu
INSERT INTO categories (name, slug, description, icon_url, sort_order) VALUES
('Lập trình', 'lap-trinh', 'Các khóa học về lập trình và phát triển phần mềm', '/icons/programming.svg', 1),
('Thiết kế', 'thiet-ke', 'Thiết kế đồ họa, UI/UX và thiết kế web', '/icons/design.svg', 2),
('Kinh doanh', 'kinh-doanh', 'Kỹ năng kinh doanh và quản lý', '/icons/business.svg', 3),
('Marketing', 'marketing', 'Digital marketing và quảng cáo', '/icons/marketing.svg', 4),
('Nhiếp ảnh', 'nhiep-anh', 'Kỹ thuật chụp ảnh và chỉnh sửa', '/icons/photography.svg', 5)
ON CONFLICT (slug) DO NOTHING;

-- Thêm danh mục con
INSERT INTO categories (name, slug, description, parent_id, sort_order) VALUES
('Web Development', 'web-development', 'HTML, CSS, JavaScript, React, Node.js', (SELECT id FROM categories WHERE slug = 'lap-trinh'), 1),
('Mobile Development', 'mobile-development', 'iOS, Android, React Native, Flutter', (SELECT id FROM categories WHERE slug = 'lap-trinh'), 2),
('Data Science', 'data-science', 'Python, R, Machine Learning, AI', (SELECT id FROM categories WHERE slug = 'lap-trinh'), 3),
('UI/UX Design', 'ui-ux-design', 'User Interface và User Experience Design', (SELECT id FROM categories WHERE slug = 'thiet-ke'), 1),
('Graphic Design', 'graphic-design', 'Adobe Photoshop, Illustrator, InDesign', (SELECT id FROM categories WHERE slug = 'thiet-ke'), 2)
ON CONFLICT (slug) DO NOTHING;

-- Thêm người dùng admin mẫu
INSERT INTO users (email, username, password_hash, first_name, last_name, role, is_verified) VALUES
('admin@toanthaycong.com', 'admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Admin', 'System', 'admin', true)
ON CONFLICT (email) DO NOTHING;

-- Thêm giảng viên mẫu
INSERT INTO users (email, username, password_hash, first_name, last_name, role, is_verified) VALUES
('instructor@toanthaycong.com', 'instructor1', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Nguyễn', 'Văn Giảng', 'instructor', true)
ON CONFLICT (email) DO NOTHING;

-- Thêm học viên mẫu
INSERT INTO users (email, username, password_hash, first_name, last_name, role, is_verified) VALUES
('student1@example.com', 'student1', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Trần', 'Văn Học', 'student', true),
('student2@example.com', 'student2', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Lê', 'Thị Sinh', 'student', true)
ON CONFLICT (email) DO NOTHING;

-- Tạo hồ sơ giảng viên
INSERT INTO instructor_profiles (user_id, title, expertise, experience_years, is_approved) VALUES
((SELECT id FROM users WHERE username = 'instructor1'), 
 'Senior Full-stack Developer', 
 ARRAY['JavaScript', 'React', 'Node.js', 'Python', 'Go'], 
 8, 
 true)
ON CONFLICT (user_id) DO NOTHING;

-- Thêm các thẻ tag mẫu
INSERT INTO tags (name, slug, description, color) VALUES
('JavaScript', 'javascript', 'JavaScript programming language', '#F7DF1E'),
('React', 'react', 'React.js library', '#61DAFB'),
('Node.js', 'nodejs', 'Node.js runtime', '#339933'),
('Python', 'python', 'Python programming language', '#3776AB'),
('Beginner', 'beginner', 'Suitable for beginners', '#28A745'),
('Advanced', 'advanced', 'For advanced learners', '#DC3545')
ON CONFLICT (slug) DO NOTHING;

-- Thêm khóa học mẫu
INSERT INTO courses (
    title, 
    slug, 
    description, 
    short_description,
    instructor_id, 
    category_id, 
    price, 
    discount_price,
    language,
    level,
    status,
    requirements,
    what_you_learn,
    target_audience,
    published_at
) VALUES (
    'Khóa học React.js từ cơ bản đến nâng cao',
    'react-js-co-ban-den-nang-cao',
    'Khóa học toàn diện về React.js, từ những kiến thức cơ bản nhất đến các kỹ thuật nâng cao. Bạn sẽ học cách xây dựng các ứng dụng web hiện đại với React.',
    'Học React.js từ A-Z với các dự án thực tế',
    (SELECT id FROM users WHERE username = 'instructor1'),
    (SELECT id FROM categories WHERE slug = 'web-development'),
    999000,
    699000,
    'vi',
    'beginner',
    'published',
    ARRAY['Kiến thức HTML, CSS cơ bản', 'JavaScript ES6+', 'Máy tính có kết nối internet'],
    ARRAY['Xây dựng ứng dụng React từ đầu', 'Quản lý state với Redux', 'React Hooks', 'Testing với Jest'],
    ARRAY['Người mới bắt đầu học React', 'Developer muốn nâng cao kỹ năng', 'Sinh viên IT'],
    CURRENT_TIMESTAMP
)
ON CONFLICT (slug) DO NOTHING;
