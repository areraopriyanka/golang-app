-- +goose Up


DELETE FROM master_user_otp
WHERE user_id IN (
  SELECT id FROM master_user_records
  WHERE mobile_no IN (
    SELECT mobile_no
    FROM master_user_records
    GROUP BY mobile_no
    HAVING COUNT(*) > 1
  )
  AND user_status IN (
    'ONBOARDING',
    'USER_CREATED',
    'AGREEMENTS_REVIEWED',
    'AGE_VERIFICATION_PASSED',
    'AGE_VERIFICATION_FAILED',
    'PHONE_VERIFICATION_OTP_SENT',
    'PHONE_NUMBER_VERIFIED',
    'ADDRESS_CONFIRMED',
    'PASSWORD_SET',
    'KYC_FAIL'
  )
);


DELETE FROM sardine_kyc_data
WHERE user_id IN (
  SELECT id FROM master_user_records
  WHERE mobile_no IN (
    SELECT mobile_no
    FROM master_user_records
    GROUP BY mobile_no
    HAVING COUNT(*) > 1
  )
  AND user_status IN (
    'ONBOARDING',
    'USER_CREATED',
    'AGREEMENTS_REVIEWED',
    'AGE_VERIFICATION_PASSED',
    'AGE_VERIFICATION_FAILED',
    'PHONE_VERIFICATION_OTP_SENT',
    'PHONE_NUMBER_VERIFIED',
    'ADDRESS_CONFIRMED',
    'PASSWORD_SET',
    'KYC_FAIL'
  )
);


DELETE FROM master_user_records
WHERE mobile_no IN (
  SELECT mobile_no
  FROM master_user_records
  GROUP BY mobile_no
  HAVING COUNT(*) > 1
)
AND user_status IN (
  'ONBOARDING'
  'USER_CREATED',
  'AGREEMENTS_REVIEWED',
  'AGE_VERIFICATION_PASSED',
  'AGE_VERIFICATION_FAILED',
  'PHONE_VERIFICATION_OTP_SENT',
  'PHONE_NUMBER_VERIFIED',
  'ADDRESS_CONFIRMED',
  'PASSWORD_SET',
  'KYC_FAIL'
);

UPDATE master_user_records
SET mobile_no = NULL
WHERE id IN (
  SELECT id
  FROM (
    SELECT id,
           ROW_NUMBER() OVER (PARTITION BY mobile_no ORDER BY created_at DESC) AS rownumber
    FROM master_user_records
    WHERE mobile_no IS NOT NULL AND user_status = 'ACTIVE'
  ) AS duplicates
  WHERE rownumber > 1
);


ALTER TABLE master_user_records ADD CONSTRAINT mobile_no_unique UNIQUE (mobile_no);