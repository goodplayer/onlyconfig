-- ===========================================web manager tables===================================================
-- management service database model:
--     independent entities: datacenter, environment, namespace
--     application => namespace
--     application [links] N namespaces
--     namespace (+ environment + datacenter) => config (multiple applications may share the same configurations)
--     org [has] application / namespace
--     org => user

create sequence onlyconfig_version_seq increment by 16 minvalue 1 maxvalue 9223372036854775807 start 1 cache 1 no cycle;

-- ---------------------------------------------------------------------------------------
-- application related metadata
-- ---------------------------------------------------------------------------------------

create table onlyconfig_datacenter
(
    datacenter_name        varchar not null,
    datacenter_description varchar not null,
    time_created           bigint  not null,
    time_updated           bigint  not null,
    primary key (datacenter_name)
);

insert into onlyconfig_datacenter (datacenter_name, datacenter_description, time_created, time_updated)
values ('default', 'default datacenter', EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000,
        EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000);

create table onlyconfig_environment
(
    env_name        varchar not null,
    env_description varchar not null,
    time_created    bigint  not null,
    time_updated    bigint  not null,
    primary key (env_name)
);

insert into onlyconfig_environment (env_name, env_description, time_created, time_updated)
VALUES ('DEV', 'production environment', EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000,
        EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),
       ('UAT', 'UAT environment', EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 + 1,
        EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 + 1),
       ('PRE', 'pre-production environment', EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 + 2,
        EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 + 2),
       ('PROD', 'development environment', EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 + 3,
        EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 + 3);

create table onlyconfig_application
(
    application_id          bigserial not null,
    application_name        varchar   not null,
    application_description varchar   not null,
    application_owner_org   varchar   not null,
    time_created            bigint    not null,
    time_updated            bigint    not null,
    primary key (application_id)
);

create unique index on onlyconfig_application (application_name);

create table onlyconfig_application_detail
(
    application_detail_id bigserial not null,
    application_id        bigint    not null,
    env_name              varchar   not null,
    datacenter_name       varchar   not null,
    time_created          bigint    not null,
    time_updated          bigint    not null,
    primary key (application_detail_id)
);

create unique index on onlyconfig_application_detail (application_id, env_name, datacenter_name);

-- ---------------------------------------------------------------------------------------
-- configuration metadata and data
-- ---------------------------------------------------------------------------------------

-- Namespaces owned by applications
-- namespace equals group in the configure api
create table onlyconfig_namespace
(
    namespace_name        varchar not null,
    namespace_description varchar not null,
    namespace_type        varchar not null,
    namespace_app         bigint  not null,
    time_created          bigint  not null,
    time_updated          bigint  not null,
    primary key (namespace_name)
);

create index on onlyconfig_namespace (namespace_app);

create index on onlyconfig_namespace ((case when namespace_type = 'public' then namespace_type end));

comment on column onlyconfig_namespace.namespace_type is 'app:"application namespace", public:"public namespace"';

-- Namespaces linked by application
-- Usage: applications uses public namespaces
create table onlyconfig_app_ns_link
(
    mapping_id     bigserial not null,
    application_id bigint    not null,
    namespace_name varchar   not null,
    time_created   bigint    not null,
    time_updated   bigint    not null,
    primary key (mapping_id)
);

create unique index on onlyconfig_app_ns_link (application_id, namespace_name);

create table onlyconfig_config
(
    config_id           bigserial not null,
    config_key          varchar   not null,
    config_namespace    varchar   not null,
    config_env          varchar   not null,
    config_datacenter   varchar   not null,
    config_content_type varchar   not null,
    config_content      varchar   not null,
    config_version      varchar   not null,
    config_status       bigint    not null,
    time_created        bigint    not null,
    time_updated        bigint    not null,
    primary key (config_id)
);

create index on onlyconfig_config (config_namespace, config_key);

create index on onlyconfig_config (config_namespace, config_env, config_datacenter, config_key);

comment on column onlyconfig_config.config_content_type is 'general:"no specific file type", json:"json file", toml:"toml file", yaml:"yaml file", properties:"properties file", xml:"xml", html:"html"';

comment on column onlyconfig_config.config_status is '0-normal, 1-deleted';

-- ---------------------------------------------------------------------------------------
-- user related metadata
-- ---------------------------------------------------------------------------------------

create table onlyconfig_org
(
    org_id       varchar not null,
    org_name     varchar not null,
    time_created bigint  not null,
    time_updated bigint  not null,
    primary key (org_id)
);

create unique index on onlyconfig_org (org_name);

insert into onlyconfig_org (org_id, org_name, time_created, time_updated)
values ('1', 'GeneralOrg', EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000);

create table onlyconfig_user
(
    user_id          varchar not null,
    username         varchar not null,
    password         varchar not null,
    display_name     varchar not null,
    email            varchar not null,
    user_status      bigint  not null,
    external_type    varchar not null,
    external_user_id varchar not null,
    time_created     bigint  not null,
    time_updated     bigint  not null,
    primary key (user_id)
);

create unique index on onlyconfig_user (username);

comment on column onlyconfig_user.user_status is '0-normal, 1-disabled';

comment on column onlyconfig_user.external_type is '(empty):"internal user", otherwise:"specific type of user source, e.g. LDAP,SSO,etc."';

insert into onlyconfig_user (user_id, username, password, display_name, email, user_status, external_type,
                             external_user_id,
                             time_created, time_updated)
values ('1', 'admin', '$2a$14$EQS3g4pLrTOdR03mKnE8i.zxbJvOYWOmZ4SjwZ.hkM0WdjENnp3sa', 'administrator',
        'example@example.com', 0, '', '',
        EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000);

create table onlyconfig_user_org_mapping
(
    user_org_mapping_id bigserial not null,
    org_id              varchar   not null,
    user_id             varchar   not null,
    role_type           int       not null,
    time_created        bigint    not null,
    time_updated        bigint    not null,
    primary key (user_org_mapping_id)
);

create unique index on onlyconfig_user_org_mapping (org_id, user_id);

create unique index on onlyconfig_user_org_mapping (user_id, org_id);

comment on column onlyconfig_user_org_mapping.role_type is '1:owner, 2:user';

insert into onlyconfig_user_org_mapping (org_id, user_id, role_type, time_created, time_updated)
values ('1', '1', 1, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000);
