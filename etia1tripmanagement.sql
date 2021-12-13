-- phpMyAdmin SQL Dump
-- version 5.1.1
-- https://www.phpmyadmin.net/
--
-- Host: 127.0.0.1
-- Generation Time: Dec 13, 2021 at 02:08 PM
-- Server version: 10.4.21-MariaDB
-- PHP Version: 8.0.11

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `etia1tripmanagement`
--

-- --------------------------------------------------------

--
-- Table structure for table `available_driver`
--

CREATE TABLE `available_driver` (
  `driverId` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- --------------------------------------------------------

--
-- Table structure for table `ongoing_trip`
--

CREATE TABLE `ongoing_trip` (
  `id` int(11) NOT NULL,
  `postalCode` varchar(127) NOT NULL,
  `passengerId` int(11) NOT NULL,
  `driverId` int(11) NOT NULL,
  `startTime` bigint(20) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- --------------------------------------------------------

--
-- Table structure for table `trip_history`
--

CREATE TABLE `trip_history` (
  `id` int(11) NOT NULL,
  `postalCode` varchar(127) NOT NULL,
  `passengerId` int(11) NOT NULL,
  `driverId` int(11) NOT NULL,
  `startTime` bigint(20) NOT NULL,
  `endTIme` bigint(20) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--
-- Indexes for dumped tables
--

--
-- Indexes for table `available_driver`
--
ALTER TABLE `available_driver`
  ADD PRIMARY KEY (`driverId`);

--
-- Indexes for table `ongoing_trip`
--
ALTER TABLE `ongoing_trip`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `userId` (`passengerId`),
  ADD UNIQUE KEY `driverId` (`driverId`);

--
-- Indexes for table `trip_history`
--
ALTER TABLE `trip_history`
  ADD PRIMARY KEY (`id`),
  ADD KEY `passengerId` (`passengerId`),
  ADD KEY `driverId` (`driverId`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `ongoing_trip`
--
ALTER TABLE `ongoing_trip`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
