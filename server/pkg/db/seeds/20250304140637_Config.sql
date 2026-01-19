-- +goose Up

--Execute below script only in dreamfi db

INSERT INTO `middleware`.`configs` (`config_name`,`type`,`value`,`user_type`,`created_by`,`updated_at`,`updated_by`)
VALUES ('CARDS_DEFAULT_CONFIG', 'JSON', '{\"product\":\"DEFAULT\",\"channel\":\"PULSE\",\"program\":\"DEFAULT\"}', 'BOTH', 'ADMIN', NULL, NULL);

INSERT INTO `middleware`.`configs` (`config_name`,`type`,`value`,`user_type`,`created_by`,`updated_at`,`updated_by`)
VALUES ('RTP_OUT_PAYMENTS_DEFAULT_CONFIG', 'JSON', '{\"product\":\"DEFAULT\",\"channel\":\"CLEARING_HOUSE\",\"type\":\"RTP_OUT\",\"program\":\"DEFAULT\"}', 'ONBOARDED', 'ADMIN', NULL, NULL);

INSERT INTO `middleware`.`configs` (`config_name`,`type`,`value`,`user_type`,`created_by`,`updated_at`,`updated_by`)
VALUES ('ACH_OUT_PAYMENTS_DEFAULT_CONFIG', 'JSON', '{\"product\":\"DEFAULT\",\"channel\":\"ACH\",\"type\":\"ACH_OUT\",\"program\":\"DEFAULT\"}', 'ONBOARDED', 'ADMIN', NULL, NULL);

INSERT INTO `middleware`.`configs` (`config_name`,`type`,`value`,`user_type`,`created_by`,`updated_at`,`updated_by`)
VALUES ('WIRE_OUT_PAYMENTS_DEFAULT_CONFIG', 'JSON', '{\"product\":\"DEFAULT\",\"channel\":\"WIRE\",\"type\":\"WIRE_OUT\",\"program\":\"DEFAULT\"}', 'ONBOARDED', 'ADMIN', NULL, NULL);

INSERT INTO `middleware`.`configs` (`config_name`,`type`,`value`,`user_type`,`created_by`,`updated_at`,`updated_by`)
VALUES ('INTERNAL_OUT_PAYMENTS_DEFAULT_CONFIG', 'JSON', '{\"product\":\"DEFAULT\",\"channel\":\"INTERNAL\",\"type\":\"INTERNAL_TRANSFER\",\"program\":\"DEFAULT\"}', 'ONBOARDED', 'ADMIN', NULL, NULL);

INSERT INTO `middleware`.`configs` (`config_name`,`type`,`value`,`user_type`,`created_by`,`updated_at`,`updated_by`)
VALUES ('PAGE_SIZE_CONFIG', 'JSON', '{\"accountsList\": 100,\"cardsList\": 100,\"accountTransactionsList\": 30,\"cardTransactionsList\": 30}', 'ONBOARDED', 'ADMIN', NULL, NULL);

INSERT INTO `middleware`.`configs` (`config_name`,`type`,`value`,`user_type`,`created_by`,`updated_at`,`updated_by`)
VALUES ('ACH_PULL_PAYMENTS_DEFAULT_CONFIG', 'JSON', '{\"product\":\"DEFAULT\",\"channel\":\"ACH\",\"type\":\"ACH_PULL\",\"program\":\"DEFAULT\"}', 'ONBOARDED', 'ADMIN', NULL, NULL);

INSERT INTO `middleware`.`configs` (`config_name`,`type`,`value`,`user_type`,`created_by`,`updated_at`,`updated_by`)
VALUES ('ADDRESS_MODE_CONFIG', 'JSON', '{\"type\":\"MANUAL\"}', 'BOTH', 'ADMIN', NULL, NULL);

INSERT INTO `middleware`.`configs` (`config_name`,`type`,`value`,`user_type`,`created_by`,`updated_at`,`updated_by`)
VALUES ('INDIVIDUAL_EXTERNAL_BANK_COUNT_CONFIG', 'JSON', '{\"count\":\"5\"}', 'ONBOARDED', 'ADMIN', NULL, NULL);

INSERT INTO `middleware`.`configs` (`config_name`,`type`,`value`,`user_type`,`created_by`,`updated_at`,`updated_by`)
VALUES ('INDIVIDUAL_GETQUESTIONAIESBYPRODUCT_CONFIG', 'JSON', '{\"name\":\"DEFAULT\"}', 'ONBOARDING', 'ADMIN', NULL, NULL);

INSERT INTO `middleware`.`configs` (`config_name`,`type`,`value`,`user_type`,`created_by`,`updated_at`,`updated_by`)
VALUES ('INDIVIDUAL_GETTERMSANDCONDTIONBYPRODUCT_CONFIG', 'JSON', '{\"name\":\"DEFAULT\"}', 'ONBOARDING', 'ADMIN', NULL, NULL);

INSERT INTO `middleware`.`configs` (`config_name`,`type`,`value`,`user_type`,`created_by`,`updated_at`,`updated_by`)
VALUES ('INDIVIDUAL_ADDACCOUNTCATEGORY_CONFIG', 'JSON', '{\"name\":\"DEFAULT\"}', 'BOTH', 'ADMIN', NULL, NULL);

INSERT INTO `middleware`.`configs` (`config_name`,`type`,`value`,`user_type`,`created_by`,`updated_at`,`updated_by`)
VALUES ('FILE_MAX_SIZE_CONFIG','INT','5120','BOTH','ADMIN',NULL,NULL);

INSERT INTO `middleware`.`configs` (`config_name`,`type`,`value`,`user_type`,`created_by`,`updated_at`,`updated_by`)
VALUES ('INDIVIDUAL_ONBOARDING_API_SEQUENCE_CONFIG','STRING_ARRAY','["KYC","ADD_CUSTOMER","UPDATE_CUSTOMER_SETTINGS","ADD_CHECKING_ACCOUNT","END_ONBOARDING"]','ONBOARDING','ADMIN',NULL,NULL);

INSERT INTO `middleware`.`configs` (`config_name`,`type`,`value`,`user_type`,`created_by`,`updated_at`,`updated_by`)
VALUES ('INDIVIDUAL_ADDSAVINGSACCOUNTCATEGORY_CONFIG', 'JSON', '{\"name\":\"HYSA\"}', 'BOTH', 'ADMIN', NULL, NULL);
