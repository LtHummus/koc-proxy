
create table users (
                            id uuid primary key not null,
                            username character varying not null,
                            secret_hash numeric,
                            banned_until timestamp without time zone,
                            discord_id character varying not null,
                            banned_reason character varying,
                            last_login timestamp without time zone default null
);
create unique index users_discord_id_uindex on users using btree (discord_id);
create unique index users_username_uindex on users using btree (username);

create table ban_records (
                                  id uuid not null,
                                  user_guid uuid not null,
                                  ban_time timestamp without time zone not null,
                                  ban_expiration timestamp without time zone not null,
                                  reporter_discord_id character varying not null,
                                  ban_reason character varying not null,
                                  foreign key (user_guid) references users (id)
                                      match simple on update no action on delete no action
);



