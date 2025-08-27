CREATE TABLE
    activity_codes (
        data_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        activity_code TEXT UNIQUE,                          
        activity_group_code TEXT,
        activity_code_explanation TEXT
    );