-- +goose Up
--
-- PostgreSQL database dump
--


-- Dumped from database version 17.5 (Debian 17.5-1.pgdg120+1)
-- Dumped by pg_dump version 18.4 (Debian 18.4-1.pgdg13+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: auditorium; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.auditorium (
    id integer NOT NULL,
    number character varying NOT NULL,
    building_id integer NOT NULL,
    rasp_id integer NOT NULL
);


--
-- Name: auditorium_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.auditorium_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: auditorium_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.auditorium_id_seq OWNED BY public.auditorium.id;


--
-- Name: building; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.building (
    id integer NOT NULL,
    letter character varying NOT NULL,
    title character varying NOT NULL
);


--
-- Name: building_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.building_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: building_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.building_id_seq OWNED BY public.building.id;


--
-- Name: department; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.department (
    id integer NOT NULL,
    title character varying NOT NULL,
    title_short character varying NOT NULL,
    faculty_id integer NOT NULL,
    rasp_id integer NOT NULL
);


--
-- Name: department_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.department_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: department_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.department_id_seq OWNED BY public.department.id;


--
-- Name: faculty; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.faculty (
    id integer NOT NULL,
    title character varying NOT NULL,
    title_short character varying NOT NULL,
    rasp_id integer NOT NULL
);


--
-- Name: faculty_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.faculty_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: faculty_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.faculty_id_seq OWNED BY public.faculty.id;


--
-- Name: group; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public."group" (
    id integer NOT NULL,
    number character varying NOT NULL,
    faculty_id integer NOT NULL,
    rasp_id integer NOT NULL,
    course integer NOT NULL
);


--
-- Name: group_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.group_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: group_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.group_id_seq OWNED BY public."group".id;


--
-- Name: lesson; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.lesson (
    id integer NOT NULL,
    date date NOT NULL,
    group_id integer NOT NULL,
    title character varying NOT NULL,
    "time" character varying NOT NULL,
    week_type character varying NOT NULL,
    start_time timestamp without time zone NOT NULL,
    end_time timestamp without time zone NOT NULL,
    type character varying
);


--
-- Name: lesson_auditorium; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.lesson_auditorium (
    id integer NOT NULL,
    lesson_id integer,
    auditorium_id integer
);


--
-- Name: lesson_auditorium_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.lesson_auditorium_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: lesson_auditorium_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.lesson_auditorium_id_seq OWNED BY public.lesson_auditorium.id;


--
-- Name: lesson_auditorium_teacher; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.lesson_auditorium_teacher (
    id integer NOT NULL,
    lesson_id integer NOT NULL,
    auditorium_id integer,
    teacher_id integer
);


--
-- Name: lesson_auditorium_teacher_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.lesson_auditorium_teacher_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: lesson_auditorium_teacher_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.lesson_auditorium_teacher_id_seq OWNED BY public.lesson_auditorium_teacher.id;


--
-- Name: lesson_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.lesson_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: lesson_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.lesson_id_seq OWNED BY public.lesson.id;


--
-- Name: lesson_teacher; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.lesson_teacher (
    id integer NOT NULL,
    lesson_id integer,
    teacher_id integer
);


--
-- Name: lesson_teacher_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.lesson_teacher_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: lesson_teacher_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.lesson_teacher_id_seq OWNED BY public.lesson_teacher.id;


--
-- Name: teacher; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.teacher (
    id integer NOT NULL,
    full_name character varying NOT NULL,
    short_name character varying NOT NULL,
    rasp_id integer,
    link character varying
);


--
-- Name: teacher_department; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.teacher_department (
    id integer NOT NULL,
    department_id integer NOT NULL,
    teacher_id integer NOT NULL
);


--
-- Name: teacher_department_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.teacher_department_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: teacher_department_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.teacher_department_id_seq OWNED BY public.teacher_department.id;


--
-- Name: teacher_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.teacher_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: teacher_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.teacher_id_seq OWNED BY public.teacher.id;


--
-- Name: auditorium id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auditorium ALTER COLUMN id SET DEFAULT nextval('public.auditorium_id_seq'::regclass);


--
-- Name: building id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.building ALTER COLUMN id SET DEFAULT nextval('public.building_id_seq'::regclass);


--
-- Name: department id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.department ALTER COLUMN id SET DEFAULT nextval('public.department_id_seq'::regclass);


--
-- Name: faculty id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.faculty ALTER COLUMN id SET DEFAULT nextval('public.faculty_id_seq'::regclass);


--
-- Name: group id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."group" ALTER COLUMN id SET DEFAULT nextval('public.group_id_seq'::regclass);


--
-- Name: lesson id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lesson ALTER COLUMN id SET DEFAULT nextval('public.lesson_id_seq'::regclass);


--
-- Name: lesson_auditorium id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lesson_auditorium ALTER COLUMN id SET DEFAULT nextval('public.lesson_auditorium_id_seq'::regclass);


--
-- Name: lesson_auditorium_teacher id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lesson_auditorium_teacher ALTER COLUMN id SET DEFAULT nextval('public.lesson_auditorium_teacher_id_seq'::regclass);


--
-- Name: lesson_teacher id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lesson_teacher ALTER COLUMN id SET DEFAULT nextval('public.lesson_teacher_id_seq'::regclass);


--
-- Name: teacher id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teacher ALTER COLUMN id SET DEFAULT nextval('public.teacher_id_seq'::regclass);


--
-- Name: teacher_department id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teacher_department ALTER COLUMN id SET DEFAULT nextval('public.teacher_department_id_seq'::regclass);


--
-- Name: auditorium auditorium_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auditorium
    ADD CONSTRAINT auditorium_pkey PRIMARY KEY (id);


--
-- Name: building building_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.building
    ADD CONSTRAINT building_pkey PRIMARY KEY (id);


--
-- Name: department department_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.department
    ADD CONSTRAINT department_pkey PRIMARY KEY (id);


--
-- Name: faculty faculty_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.faculty
    ADD CONSTRAINT faculty_pkey PRIMARY KEY (id);


--
-- Name: group group_number_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."group"
    ADD CONSTRAINT group_number_key UNIQUE (number);


--
-- Name: group group_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."group"
    ADD CONSTRAINT group_pkey PRIMARY KEY (id);


--
-- Name: lesson_auditorium lesson_auditorium_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lesson_auditorium
    ADD CONSTRAINT lesson_auditorium_pkey PRIMARY KEY (id);


--
-- Name: lesson_auditorium_teacher lesson_auditorium_teacher_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lesson_auditorium_teacher
    ADD CONSTRAINT lesson_auditorium_teacher_pkey PRIMARY KEY (id);


--
-- Name: lesson lesson_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lesson
    ADD CONSTRAINT lesson_pkey PRIMARY KEY (id);


--
-- Name: lesson_teacher lesson_teacher_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lesson_teacher
    ADD CONSTRAINT lesson_teacher_pkey PRIMARY KEY (id);


--
-- Name: teacher_department teacher_department_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teacher_department
    ADD CONSTRAINT teacher_department_pkey PRIMARY KEY (id);


--
-- Name: teacher teacher_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teacher
    ADD CONSTRAINT teacher_pkey PRIMARY KEY (id);


--
-- Name: idx_auditorium_number; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_auditorium_number ON public.auditorium USING btree (number);


--
-- Name: idx_faculty_title_short; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_faculty_title_short ON public.faculty USING btree (title_short);


--
-- Name: idx_group_number; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_group_number ON public."group" USING btree (number);


--
-- Name: idx_lesson_auditorium_teacher_auditorium; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lesson_auditorium_teacher_auditorium ON public.lesson_auditorium_teacher USING btree (auditorium_id);


--
-- Name: idx_lesson_auditorium_teacher_lesson; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lesson_auditorium_teacher_lesson ON public.lesson_auditorium_teacher USING btree (lesson_id);


--
-- Name: idx_lesson_auditorium_teacher_teacher; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lesson_auditorium_teacher_teacher ON public.lesson_auditorium_teacher USING btree (teacher_id);


--
-- Name: idx_lesson_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lesson_date ON public.lesson USING btree (date);


--
-- Name: idx_lesson_date_group; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lesson_date_group ON public.lesson USING btree (date, group_id);


--
-- Name: idx_teacher_department_department; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_teacher_department_department ON public.teacher_department USING btree (department_id);


--
-- Name: idx_teacher_department_teacher; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_teacher_department_teacher ON public.teacher_department USING btree (teacher_id);


--
-- Name: idx_teacher_full_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_teacher_full_name ON public.teacher USING btree (full_name);


--
-- Name: idx_teacher_short_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_teacher_short_name ON public.teacher USING btree (short_name);


--
-- Name: auditorium auditorium_building_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auditorium
    ADD CONSTRAINT auditorium_building_id_fkey FOREIGN KEY (building_id) REFERENCES public.building(id) ON DELETE CASCADE;


--
-- Name: department department_faculty_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.department
    ADD CONSTRAINT department_faculty_id_fkey FOREIGN KEY (faculty_id) REFERENCES public.faculty(id) ON DELETE CASCADE;


--
-- Name: group group_faculty_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."group"
    ADD CONSTRAINT group_faculty_id_fkey FOREIGN KEY (faculty_id) REFERENCES public.faculty(id) ON DELETE CASCADE;


--
-- Name: lesson_auditorium_teacher lesson_auditorium_teacher_auditorium_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lesson_auditorium_teacher
    ADD CONSTRAINT lesson_auditorium_teacher_auditorium_id_fkey FOREIGN KEY (auditorium_id) REFERENCES public.auditorium(id) ON DELETE SET NULL;


--
-- Name: lesson_auditorium_teacher lesson_auditorium_teacher_lesson_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lesson_auditorium_teacher
    ADD CONSTRAINT lesson_auditorium_teacher_lesson_id_fkey FOREIGN KEY (lesson_id) REFERENCES public.lesson(id) ON DELETE CASCADE;


--
-- Name: lesson_auditorium_teacher lesson_auditorium_teacher_teacher_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lesson_auditorium_teacher
    ADD CONSTRAINT lesson_auditorium_teacher_teacher_id_fkey FOREIGN KEY (teacher_id) REFERENCES public.teacher(id) ON DELETE SET NULL;


--
-- Name: lesson lesson_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lesson
    ADD CONSTRAINT lesson_group_id_fkey FOREIGN KEY (group_id) REFERENCES public."group"(id) ON DELETE CASCADE;


--
-- Name: teacher_department teacher_department_department_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teacher_department
    ADD CONSTRAINT teacher_department_department_id_fkey FOREIGN KEY (department_id) REFERENCES public.department(id) ON DELETE CASCADE;


--
-- Name: teacher_department teacher_department_teacher_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teacher_department
    ADD CONSTRAINT teacher_department_teacher_id_fkey FOREIGN KEY (teacher_id) REFERENCES public.teacher(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--
