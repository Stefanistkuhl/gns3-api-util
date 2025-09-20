create table clusters(
	cluster_id integer primary key autoincrement,
	name text not null unique,
	description text
);

create table nodes (
	node_id integer primary key autoincrement,
	cluster_id integer not null,
	protocol text check (protocol in ('http','https')) not null,
	auth_user text not null,
	host text not null,
	port integer not null,
	weight integer not null default 5 check (weight between 0 and 10),
	max_groups integer default 3,
	unique(protocol, host, port)
);

create table class(
	class_id integer primary key autoincrement,
	cluster_id integer not null,
	name text not null,
	description text,
	foreign key (cluster_id) references clusters(cluster_id) on delete cascade
);

create table groups(
	group_id integer primary key autoincrement,
	class_id integer not null,
	name text not null,
	foreign key (class_id) references class(class_id) on delete cascade
);

create table users (
	user_id integer primary key autoincrement,
	username text not null unique,
	full_name text,
	group_id integer not null,
	default_password text not null;
	foreign key (group_id) references groups(group_id) on delete cascade
);

create table group_assignments (
	group_id integer not null,
	node_id integer not null,
	assigned_at timestamp default current_timestamp,
	primary key (group_id),
	foreign key (group_id) references groups(group_id) on delete cascade,
	foreign key (node_id) references nodes(node_id) on delete cascade
);

create table exercises (
	exercise_id integer primary key autoincrement,
	group_id integer not null,
	name text not null,
	state text check (state in ('created','running','completed','deleted')) default 'created',
	created_at timestamp default current_timestamp,
	foreign key (group_id) references groups(group_id) on delete cascade
);
