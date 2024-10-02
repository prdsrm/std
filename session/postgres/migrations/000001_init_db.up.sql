-- Database initialization
CREATE TABLE bots (
    phone_number character varying(15) NOT NULL,
    user_id bigint NOT NULL,
    username character varying(32) DEFAULT ''::character varying NOT NULL,
    password character varying(64) DEFAULT ''::character varying NOT NULL,
    title character varying(64) DEFAULT 'A'::character varying NOT NULL,
    banned boolean DEFAULT false,
    creation_date date DEFAULT (CURRENT_DATE - 30) NOT NULL,
    warming boolean DEFAULT false,
    license character varying(36),
    premium boolean DEFAULT false
);

--
-- Name: bots bots_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY bots
    ADD CONSTRAINT bots_pkey PRIMARY KEY (user_id);

--
-- Name: bots_entities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE bots_entities (
    entity_id_1 bigint NOT NULL,
    entity_id_2 bigint NOT NULL,
    hash bigint
);

--
-- Name: bots_entities entities_ids_uniques; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY bots_entities
    ADD CONSTRAINT entities_ids_uniques UNIQUE (entity_id_1, entity_id_2);

--
-- Name: bots_entities entities_fk_1; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY bots_entities
    ADD CONSTRAINT entities_fk_1 FOREIGN KEY (entity_id_1) REFERENCES bots(user_id) ON DELETE CASCADE;

CREATE TABLE devices (
    bot_user_id bigint NOT NULL,
    api_id integer DEFAULT 2040 NOT NULL,
    api_hash character varying(32) DEFAULT 'b18441a1ff607e10a989891a5462e627'::character varying NOT NULL,
    session_string character varying(355) NOT NULL,
    device_model character varying(512) DEFAULT 'XPS 13 9370'::character varying NOT NULL,
    system_version character varying(32) DEFAULT 'Windows 10'::character varying NOT NULL,
    app_version character varying(32) DEFAULT '5.0.1 x64'::character varying NOT NULL,
    lang_pack character varying(16) DEFAULT 'tdesktop'::character varying NOT NULL,
    lang_code character varying(16) DEFAULT 'en'::character varying NOT NULL,
    system_lang_code character varying(16) DEFAULT 'en-US'::character varying NOT NULL,
    proxy character varying(256) DEFAULT ''::character varying NOT NULL,
    creation_date date DEFAULT CURRENT_DATE NOT NULL
);

--
-- Name: devices device_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY devices
    ADD CONSTRAINT device_unique UNIQUE (session_string);

--
-- Name: devices device_fk_1; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY devices
    ADD CONSTRAINT device_fk_1 FOREIGN KEY (bot_user_id) REFERENCES bots(user_id) ON DELETE CASCADE;
