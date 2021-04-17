INSERT INTO public.users(
	first_name, last_name, email, 
  password, access_level, created_at, updated_at)
	VALUES 
  ('Jackie', 'Sharpe', 'jackiesharp@whitehouse.com', 
   '$2a$10$EV3lwG6GrkG1XhCxXvDEbO/EAbm0IKPsTMXXUMuCRK1GyrOvh4Jpm', '3', NOW(), NOW());