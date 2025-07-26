require "spec"
require "db"
require "sqlite3"
require "../src/db"

# Helper to create a fresh database for each test
def with_test_db
  File.tempfile("test", ".db") do |file|
    db_path = file.path
    
    # Create schema
    DB.open("sqlite3://#{db_path}") do |db|
      schema = File.read("schema.sql")
      # SQLite doesn't handle multiple statements well, so split them
      statements = schema.split(";").map(&.strip).reject(&.empty?)
      statements.each do |statement|
        db.exec(statement + ";")
      end
    end
    
    # Run tests with the database
    DB.open("sqlite3://#{db_path}") do |db|
      yield db
    end
  end
end

describe "SqlcGenCrystal Integration Tests" do
  describe "User Operations" do
    it "creates and retrieves a user" do
      with_test_db do |db|
        queries = TestDb::Queries.new(db)
        
        # Create a user
        user = queries.create_user("test@example.com", "Test User", true, 25)
        user.should_not be_nil
        
        if user
          user.email.should eq("test@example.com")
          user.name.should eq("Test User")
          user.age.should eq(25)
          user.active.should be_true
          user.id.should be > 0
          
          # Get the user by ID
          fetched = queries.get_user(user.id)
          fetched.should_not be_nil
          
          if fetched
            fetched.email.should eq(user.email)
            fetched.name.should eq(user.name)
          end
        end
      end
    end
    
    it "finds user by email" do
      with_test_db do |db|
        queries = TestDb::Queries.new(db)
        
        # Create a user
        user = queries.create_user("unique@example.com", "Unique User", true, 30)
        
        # Find by email
        found = queries.get_user_by_email("unique@example.com")
        found.should_not be_nil
        
        if found && user
          found.id.should eq(user.id)
          found.name.should eq("Unique User")
        end
        
        # Non-existent email
        not_found = queries.get_user_by_email("notfound@example.com")
        not_found.should be_nil
      end
    end
    
    it "lists users" do
      with_test_db do |db|
        queries = TestDb::Queries.new(db)
        
        # Create multiple users
        queries.create_user("user1@example.com", "User One", true, 20)
        queries.create_user("user2@example.com", "User Two", false, 25)
        queries.create_user("user3@example.com", "User Three", true, 30)
        
        # List all users
        all_users = queries.list_users
        all_users.size.should eq(3)
        
        # List active users
        active_users = queries.list_active_users
        active_users.size.should eq(2)
        active_users.all? { |u| u.active }.should be_true
      end
    end
    
    it "updates a user" do
      with_test_db do |db|
        queries = TestDb::Queries.new(db)
        
        # Create a user
        user = queries.create_user("update@example.com", "Original Name", true, 25)
        
        if user
          # Update the user
          queries.update_user("Updated Name", false, user.id, 26)
          
          # Verify update
          updated = queries.get_user(user.id)
          if updated
            updated.name.should eq("Updated Name")
            updated.age.should eq(26)
            updated.active.should be_false
            updated.email.should eq("update@example.com") # Should not change
          end
        end
      end
    end
    
    it "deletes a user" do
      with_test_db do |db|
        queries = TestDb::Queries.new(db)
        
        # Create a user
        user = queries.create_user("delete@example.com", "To Delete", true, 25)
        
        if user
          # Delete the user
          queries.delete_user(user.id)
          
          # Verify deletion
          deleted = queries.get_user(user.id)
          deleted.should be_nil
        end
      end
    end
    
    it "counts users" do
      with_test_db do |db|
        queries = TestDb::Queries.new(db)
        
        # Initially empty
        count = queries.count_users
        count.should eq(0)
        
        # Add users
        queries.create_user("count1@example.com", "User 1", true, 20)
        queries.create_user("count2@example.com", "User 2", true, 25)
        
        # Count again
        count = queries.count_users
        count.should eq(2)
      end
    end
    
    it "gets user statistics" do
      with_test_db do |db|
        queries = TestDb::Queries.new(db)
        
        # Create users with different ages and active status
        queries.create_user("stats1@example.com", "User 1", true, 20)
        queries.create_user("stats2@example.com", "User 2", true, 30)
        queries.create_user("stats3@example.com", "User 3", false, 40)
        
        # Get stats
        stats = queries.get_user_stats
        if stats
          stats.total_users.should eq(3)
          stats.active_users.should eq(2)
          stats.average_age.should eq(30.0)
        end
      end
    end
  end
  
  describe "Post Operations" do
    it "creates and retrieves posts" do
      with_test_db do |db|
        queries = TestDb::Queries.new(db)
        
        # Create a user first
        user = queries.create_user("author@example.com", "Author", true, 30)
        
        if user
          # Create a post
          post = queries.create_post(user.id, "Test Post", false, "This is test content", 4.5)
          post.should_not be_nil
          
          if post
            post.title.should eq("Test Post")
            post.content.should eq("This is test content")
            post.published.should be_false
            post.rating.should eq(4.5)
            post.view_count.should eq(0)
            
            # Get post with author info
            fetched = queries.get_post(post.id)
            if fetched
              fetched.title.should eq(post.title)
              fetched.author_name.should eq("Author")
              fetched.author_email.should eq("author@example.com")
            end
          end
        end
      end
    end
    
    it "lists posts by user" do
      with_test_db do |db|
        queries = TestDb::Queries.new(db)
        
        # Create users
        user1 = queries.create_user("user1@example.com", "User 1", true, 25)
        user2 = queries.create_user("user2@example.com", "User 2", true, 30)
        
        if user1 && user2
          # Create posts for user1
          queries.create_post(user1.id, "Post 1", true, "Content 1", 4.0)
          queries.create_post(user1.id, "Post 2", false, "Content 2", 3.5)
          
          # Create post for user2
          queries.create_post(user2.id, "Post 3", true, "Content 3", 4.5)
          
          # List posts for user1
          user1_posts = queries.list_posts_by_user(user1.id)
          user1_posts.size.should eq(2)
          
          # List posts for user2
          user2_posts = queries.list_posts_by_user(user2.id)
          user2_posts.size.should eq(1)
        end
      end
    end
    
    it "updates post view count" do
      with_test_db do |db|
        queries = TestDb::Queries.new(db)
        
        # Create a user and post
        user = queries.create_user("viewer@example.com", "Viewer", true, 25)
        if user
          post = queries.create_post(user.id, "Popular Post", true, "Content", 5.0)
          
          if post
            # Initial view count should be 0
            post.view_count.should eq(0)
            
            # Update view count
            rows_affected = queries.update_post_view_count(post.id)
            rows_affected.should eq(1)
            
            # Check updated view count
            updated = queries.get_post(post.id)
            if updated
              updated.view_count.should eq(1)
            end
          end
        end
      end
    end
    
    it "searches posts" do
      with_test_db do |db|
        queries = TestDb::Queries.new(db)
        
        # Create a user
        user = queries.create_user("searcher@example.com", "Searcher", true, 25)
        
        if user
          # Create posts with different content
          queries.create_post(user.id, "Crystal Programming", true, "Learn Crystal language", 5.0)
          queries.create_post(user.id, "Ruby Tutorial", true, "Ruby programming guide", 4.0)
          queries.create_post(user.id, "Private Post", false, "This contains Crystal info", 3.0)
          
          # Search for "Crystal" (should find only published posts)
          results = queries.search_posts("%Crystal%", "%Crystal%")
          results.size.should eq(1)
          results[0].title.should eq("Crystal Programming")
          
          # Search for "programming"
          results = queries.search_posts("%programming%", "%programming%")
          results.size.should eq(2)
        end
      end
    end
  end
  
  describe "Tag Operations" do
    it "creates tags and associates with posts" do
      with_test_db do |db|
        queries = TestDb::Queries.new(db)
        
        # Create a user and post
        user = queries.create_user("tagger@example.com", "Tagger", true, 25)
        if user
          post = queries.create_post(user.id, "Tagged Post", true, "Content", 4.0)
          
          if post
            # Create tags
            tag1 = queries.create_tag("crystal")
            tag2 = queries.create_tag("programming")
            
            if tag1 && tag2
              # Associate tags with post
              queries.add_post_tag(post.id, tag1.id)
              queries.add_post_tag(post.id, tag2.id)
              
              # Get post tags
              tags = queries.get_post_tags(post.id)
              tags.size.should eq(2)
              tags.map(&.name).should contain("crystal")
              tags.map(&.name).should contain("programming")
            end
          end
        end
      end
    end
  end
  
  describe "Type Mappings" do
    it "handles nullable columns correctly" do
      with_test_db do |db|
        queries = TestDb::Queries.new(db)
        
        # Create user with null age
        user = queries.create_user("nullable@example.com", "Nullable User", true, nil)
        user.should_not be_nil
        
        if user
          user.age.should be_nil
          
          # Retrieve and verify
          fetched = queries.get_user(user.id)
          if fetched
            fetched.age.should be_nil
          end
        end
      end
    end
    
    it "handles boolean values correctly" do
      with_test_db do |db|
        queries = TestDb::Queries.new(db)
        
        # Create active and inactive users
        active_user = queries.create_user("active@example.com", "Active", true, 25)
        inactive_user = queries.create_user("inactive@example.com", "Inactive", false, 30)
        
        if active_user && inactive_user
          active_user.active.should be_true
          inactive_user.active.should be_false
        end
      end
    end
    
    it "handles datetime values correctly" do
      with_test_db do |db|
        queries = TestDb::Queries.new(db)
        
        # Create a user
        user = queries.create_user("time@example.com", "Time User", true, 25)
        
        if user
          # created_at should be set
          user.created_at.should_not be_nil
          user.created_at.should be_a(Time)
          
          # Should be recent (within last minute)
          (Time.utc - user.created_at).total_seconds.should be < 60
        end
      end
    end
    
    it "handles real/float values correctly" do
      with_test_db do |db|
        queries = TestDb::Queries.new(db)
        
        # Create user and post with rating
        user = queries.create_user("rater@example.com", "Rater", true, 25)
        if user
          post = queries.create_post(user.id, "Rated Post", true, "Content", 4.75)
          
          if post
            post.rating.should eq(4.75)
            post.rating.should be_a(Float64)
          end
        end
      end
    end
  end
end